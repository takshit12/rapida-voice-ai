# Voice Pipeline Architecture

## Overview

Rapida Voice AI implements a real-time voice orchestration platform that connects phone callers to AI assistants via Twilio media streams. The system handles speech-to-text (STT), LLM processing, and text-to-speech (TTS) in a streaming pipeline.

---

## End-to-End Call Flow

```
Caller
  │
  ▼
Twilio (Phone Network)
  │
  ├─ 1. Webhook: GET /v1/talk/twilio/call/:assistantId
  │     → CallReciever() creates conversation, returns TwiML
  │
  ├─ 2. TwiML instructs Twilio to open WebSocket:
  │     wss://<PUBLIC_ASSISTANT_HOST>/v1/talk/twilio/usr/...
  │
  ├─ 3. WebSocket: CallTalker() upgrades connection
  │     → Twilio streams mulaw audio (8kHz, mono, base64)
  │
  ▼
Assistant API (port 9007)
  │
  ├─ Audio Input Pipeline (connectMicrophone)
  │   ├─ Denoiser (RNNoise) → removes background noise
  │   ├─ VAD (Silero) → detects voice activity / interruptions
  │   ├─ STT (Deepgram/Google/Azure) → transcribes speech to text
  │   └─ End-of-Speech → detects when user stops speaking
  │
  ├─ LLM Processing
  │   ├─ Conversation history assembled
  │   ├─ LLM generates response (OpenAI/Anthropic/Groq/etc.)
  │   └─ Tool calls executed if needed
  │
  ├─ Audio Output Pipeline (connectSpeaker)
  │   ├─ Text Aggregator → assembles sentences from LLM stream
  │   └─ TTS (ElevenLabs/Google/Azure) → converts text to audio
  │
  ▼
Twilio WebSocket (audio sent back)
  │
  ▼
Caller hears response
```

---

## Key Components

### Telephony Layer

**Files:**
- `api/assistant-api/api/talk/inbound-call.go` - HTTP endpoints for Twilio webhooks
- `api/assistant-api/internal/telephony/internal/twilio/telephony.go` - TwiML generation
- `api/assistant-api/internal/telephony/internal/twilio/websocket.go` - WebSocket media streaming

**Supported Providers:** Twilio, Vonage, Exotel

**Twilio WebSocket Protocol:**

| Event | Direction | Description |
|-------|-----------|-------------|
| `connected` | Twilio → Server | WebSocket established |
| `start` | Twilio → Server | Stream started, contains `streamSid` |
| `media` | Twilio → Server | Audio chunk (base64 mulaw) |
| `media` | Server → Twilio | Audio response (base64 mulaw) |
| `clear` | Server → Twilio | Interrupt/clear audio buffer |
| `stop` | Twilio → Server | Stream ended |

### Audio Configuration

**File:** `api/assistant-api/internal/audio/config.go`

| Format | Sample Rate | Encoding | Use Case |
|--------|-------------|----------|----------|
| Mulaw 8kHz Mono | 8000 Hz | mu-law | Twilio telephony |
| Linear 8kHz Mono | 8000 Hz | PCM | Alternative telephony |
| Linear 16kHz Mono | 16000 Hz | PCM | VAD processing, high-quality STT |
| Linear 24kHz Mono | 24000 Hz | PCM | High-quality TTS |

### Voice Activity Detection (VAD)

**Files:**
- `api/assistant-api/internal/vad/vad.go` - VAD factory
- `api/assistant-api/internal/vad/internal/silero_vad/silero_vad.go` - Silero VAD implementation

**How it works:**
1. Audio from Twilio arrives as mulaw 8kHz
2. Resampled to linear PCM 16kHz (required by Silero)
3. Converted to float32 samples
4. Silero ONNX model detects speech segments
5. On detection, fires `InterruptionPacket` to interrupt assistant speech

**Model Resolution:**
- Checks `SILERO_MODEL_PATH` environment variable first
- Falls back to `runtime.Caller(0)` relative path (development only)
- In production Docker: `SILERO_MODEL_PATH=/opt/apps/models/silero_vad.onnx`

### Generic Adapter (Packet Router)

**Files:**
- `api/assistant-api/internal/adapters/generic/io.go` - Microphone/speaker initialization
- `api/assistant-api/internal/adapters/generic/callback_generic.go` - Packet dispatch to subsystems

**Packet Types and Flow:**

```
UserAudioPacket     → Denoiser → Recording + VAD + STT
UserTextPacket      → EndOfSpeech analysis → Message creation
SpeechToTextPacket  → EndOfSpeech → Message creation → Notify UI
EndOfSpeechPacket   → Create message → Execute LLM
InterruptionPacket  → EndOfSpeech + Recording + Notify interruption
LLMStreamPacket     → Text Aggregator → TTS
LLMMessagePacket    → Create message → Text Aggregator → TTS
TTS AudioPacket     → Recording + Notify (send audio to caller)
```

**Error Handling:** Each subsystem (VAD, STT, denoiser) is nil-checked before processing. If a subsystem fails to initialize, it's skipped without breaking the pipeline.

### Session Lifecycle

**File:** `api/assistant-api/internal/adapters/generic/session_generic.go`

```
Connect()
  ├─ Load assistant configuration
  ├─ Initialize tools
  ├─ connectMicrophone() → STT, VAD, Denoiser, EndOfSpeech
  ├─ connectSpeaker() → TTS, Text Aggregator
  └─ Initialize LLM executor

Disconnect()
  ├─ disconnectMicrophone() → Close STT, EndOfSpeech, VAD
  ├─ disconnectSpeaker() → Close TTS, Aggregator
  ├─ Save recording
  ├─ Update conversation metrics
  └─ Export to OpenSearch
```

---

## Data Model

### Conversations

```
AssistantConversation
  ├─ assistantId          → Which assistant handled the call
  ├─ source               → "phone", "web", "api"
  ├─ metadata             → Call duration, provider info
  ├─ metrics              → Token counts, latency
  ├─ telephonyEvents      → Twilio status callbacks
  └─ messages[]           → Transcript entries
```

### Messages

```
AssistantConversationMessage
  ├─ role                 → "user" or "assistant"
  ├─ body                 → Transcribed text / LLM response
  ├─ source               → "phone", "web", "api"
  ├─ metrics              → Duration, tokens, latency
  └─ metadatas            → Provider info, timestamps
```

---

## Infrastructure

### Railway Services

| Service | Port | Purpose |
|---------|------|---------|
| `assistant-api` | 9007 | Voice AI engine (WebSocket, STT, TTS, LLM) |
| `web-api` | 9001 | User management, project config, gRPC-Web API |
| `integration-api` | 9004 | External provider integrations (OpenAI, Groq, etc.) |
| `endpoint-api` | 9005 | API endpoint management |
| `document-api` | 9010 | Document processing |
| `ui` | 3000 | React frontend (served via nginx) |
| `Postgres` | 5432 | Primary database |
| `Redis` | 6379 | Session cache, authentication |

### Nginx Routing

**File:** `nginx/nginx.conf`

```
/api/assistant/*     → assistant-api:9007 (gRPC-Web + WebSocket)
/api/web/*           → web-api:9001
/api/endpoint/*      → endpoint-api:9005
/                    → ui:3000 (static files)
```

WebSocket upgrade headers are configured for the assistant-api upstream to support Twilio media streams and gRPC-Web streaming.

### Docker Build (assistant-api)

**File:** `docker/assistant-api/Dockerfile`

Multi-stage build with CGO dependencies:

| Dependency | Purpose |
|------------|---------|
| ONNX Runtime v1.16.0 | ML model inference (VAD) |
| RNNoise | Audio noise suppression |
| Azure Speech SDK | Azure STT/TTS provider |
| Silero VAD model | Voice activity detection ONNX model |

Runtime container runs as non-root `rapida-app` user from `/opt/apps`.

---

## Environment Variables (assistant-api)

**File:** `docker/assistant-api/.assistant.env`

| Variable | Description |
|----------|-------------|
| `PUBLIC_ASSISTANT_HOST` | Public domain for Twilio WebSocket URLs |
| `SILERO_MODEL_PATH` | Path to Silero VAD ONNX model |
| `ASSET_STORE__STORAGE_PATH_PREFIX` | Local file storage for recordings |
| `POSTGRES__HOST` | Database connection |
| `REDIS__HOST` | Cache connection |
| `WEB_HOST` | Internal web-api address |
| `INTEGRATION_HOST` | Internal integration-api address |

---

## Supported Providers

### Speech-to-Text (STT)
- Deepgram
- Google Cloud Speech
- Azure Cognitive Services

### Text-to-Speech (TTS)
- ElevenLabs
- Google Cloud Text-to-Speech
- Azure Cognitive Services

### Large Language Models (LLM)
- OpenAI (GPT-4, GPT-4o, etc.)
- Anthropic (Claude)
- Groq (Llama, Mixtral, Gemma)
- Google (Gemini)
