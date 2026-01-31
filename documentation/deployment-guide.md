# Deployment Guide

## Railway Deployment

Rapida Voice AI is deployed on Railway with automatic deploys from the `main` branch. Each push to `main` triggers a build and deployment.

---

## Services

The application consists of 8 Railway services:

| Service | Build Method | Dockerfile |
|---------|-------------|------------|
| `assistant-api` | Dockerfile | `docker/assistant-api/Dockerfile` |
| `web-api` | Dockerfile | `docker/web-api/Dockerfile` |
| `integration-api` | Dockerfile | `docker/integration-api/Dockerfile` |
| `endpoint-api` | Dockerfile | `docker/endpoint-api/Dockerfile` |
| `document-api` | Dockerfile | `docker/document-api/Dockerfile` |
| `ui` | Dockerfile | `docker/ui/Dockerfile` |
| `Postgres` | Railway Template | - |
| `Redis` | Railway Template | - |

---

## assistant-api Docker Build

The assistant-api has the most complex build due to CGO dependencies for audio processing.

### Builder Stage

1. **Base image:** `golang:1.25-bookworm`
2. **System deps:** gcc, g++, make, autoconf, automake, libtool, pkg-config
3. **Go modules:** Downloaded and cached
4. **ONNX Runtime v1.16.0:** ML inference engine for VAD
5. **RNNoise:** Audio noise suppression (built from source)
6. **Azure Speech SDK:** Azure STT/TTS support
7. **Silero VAD model:** Downloaded from official silero-vad repository
8. **Go binary:** Compiled with CGO flags for all native libraries

### Runtime Stage

1. **Base image:** `golang:1.25-bookworm` (minimal)
2. **User:** `rapida-app` (UID 1000, non-root)
3. **Working directory:** `/opt/apps`
4. **Copied artifacts:**
   - Compiled binary
   - ONNX Runtime libraries
   - Azure Speech SDK libraries
   - RNNoise library
   - Silero VAD ONNX model
   - Database migrations
   - Environment config

### Key Paths in Container

| Path | Contents |
|------|----------|
| `/opt/apps/assistant-api` | Compiled binary |
| `/opt/apps/models/silero_vad.onnx` | VAD model |
| `/opt/apps/env/.assistant.env` | Environment config |
| `/opt/apps/api/assistant-api/migrations/` | DB migrations |
| `/opt/apps/rapida-data/assets/workflow/` | Recording storage |
| `/opt/onnxruntime/` | ONNX Runtime libraries |
| `/opt/azure-speech-sdk/` | Azure Speech SDK |

### Environment Variables

Set via `docker/assistant-api/.assistant.env` (copied into container) and Railway service variables.

**Critical variables for phone calls:**

```env
# The public domain Twilio uses to connect WebSocket streams
# Must be the actual Railway domain, not a development placeholder
PUBLIC_ASSISTANT_HOST=assistant-api-production-2ea9.up.railway.app

# Path to Silero VAD ONNX model inside the container
SILERO_MODEL_PATH=/opt/apps/models/silero_vad.onnx

# File storage for call recordings
ASSET_STORE__STORAGE_PATH_PREFIX=/opt/apps/rapida-data/assets/workflow
```

**Internal service networking (Railway private networking):**

```env
WEB_HOST=web-api.railway.internal:9001
INTEGRATION_HOST=integration-api.railway.internal:9004
ENDPOINT_HOST=endpoint-api.railway.internal:9005
```

---

## Twilio Configuration

### Phone Number Setup

1. Purchase a phone number in the Twilio console
2. Configure the webhook URL for incoming calls:
   ```
   https://assistant-api-production-2ea9.up.railway.app/v1/talk/twilio/call/<assistantId>
   ```
3. The assistant-api will return TwiML that instructs Twilio to open a media stream

### How Twilio Connects

```
1. Caller dials phone number
2. Twilio sends webhook to assistant-api
3. assistant-api returns TwiML with <Stream> element:
   <Response>
     <Connect>
       <Stream url="wss://assistant-api-production-2ea9.up.railway.app/v1/talk/twilio/usr/...">
         <Parameter name="assistant_id" value="..." />
         <Parameter name="client_number" value="..." />
       </Stream>
     </Connect>
   </Response>
4. Twilio opens WebSocket connection to the Stream URL
5. Audio flows bidirectionally over WebSocket
```

---

## Common Issues and Fixes

### VAD Model Not Found

**Error:**
```
failed to create silero detector: failed to create session: Load model from ... failed. File doesn't exist
```

**Cause:** The Silero VAD ONNX model is not included in git (excluded by `.gitignore`). It must be downloaded during Docker build.

**Fix:** Ensure the Dockerfile downloads the model:
```dockerfile
RUN mkdir -p /opt/models && \
    wget -q -O /opt/models/silero_vad.onnx \
    https://github.com/snakers4/silero-vad/raw/master/src/silero_vad/data/silero_vad.onnx
```
And `SILERO_MODEL_PATH` is set to point to it.

### File Storage Permission Denied

**Error:**
```
unable to create complete path, err mkdir /app: permission denied
```

**Cause:** Storage path prefix points to a directory outside the runtime working directory. The non-root user cannot create arbitrary directories.

**Fix:** Set `ASSET_STORE__STORAGE_PATH_PREFIX` to a path under `/opt/apps/`.

### WebSocket Connection Fails

**Error:**
```
Websocket transport constructed with non-https:// or http:// host
```

**Cause:** The UI production config uses relative paths instead of full URLs. The gRPC-Web library requires full URLs to convert `https://` to `wss://`.

**Fix:** Use full URLs in `ui/src/configs/config.production.json`.

### Twilio Calls Connect but No Audio

**Error:** Calls show as connected in Twilio but conversation has no messages.

**Cause:** `PUBLIC_ASSISTANT_HOST` is set to an unreachable domain. Twilio can't connect the WebSocket media stream.

**Fix:** Set `PUBLIC_ASSISTANT_HOST` to the actual Railway public domain for assistant-api.

### OpenSearch Connection Error

**Error:**
```
dial tcp: lookup opensearch on ...: no such host
```

**Cause:** No OpenSearch service deployed on Railway.

**Impact:** Non-blocking. Only affects conversation search indexing. Voice calls work without it.

---

## Monitoring

### Checking Deployment Logs

Use Railway CLI or MCP tools:
```
railway logs --service assistant-api
```

**Healthy startup indicators:**
- `Listening and serving HTTP on listener what's bind with address@[::]:9007`
- `listen.initializeVAD took Xms` (VAD loaded successfully)
- `deepgram-stt: connection established` (STT connected on call)

**Error indicators:**
- `failed to create silero detector` - VAD model missing
- `illegal input audio transformer` - Audio pipeline initialization failed
- `mkdir /app: permission denied` - Storage path misconfigured
