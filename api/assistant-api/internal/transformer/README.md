# Transformer Documentation Index

Welcome to the Rapida Voice AI Transformer Package documentation. This guide will help you add support for new speech-to-text and text-to-speech providers.

## 📚 Documentation Files

### 1. [QUICKSTART.md](QUICKSTART.md) - Start Here! 🚀

**5-minute overview for experienced developers**

- Quick 5-step guide to add a new provider
- Common patterns and code snippets
- File structure template
- Credential configuration examples
- Quick reference for all existing providers

### 2. [DEVELOPMENT.md](DEVELOPMENT.md) - Comprehensive Guide

**Complete step-by-step implementation guide**

- Architecture overview
- Interface definitions with detailed explanations
- Full Step-by-Step Guide (6 steps)
- Complete code templates for STT and TTS
- Testing guidelines with examples
- Best practices and patterns
- Troubleshooting section

### 3. [ARCHITECTURE.md](ARCHITECTURE.md) - Deep Dive 🏗️

**Visual diagrams and detailed design patterns**

- System architecture diagrams
- Sequence diagrams (STT/TTS flows)
- Component structure
- Concurrency model and thread safety
- Error handling flow
- State machine diagrams
- Resource lifecycle
- Configuration hierarchy

### 4. [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md) - Implementation Checklist ✅

**Comprehensive checklist for production-ready implementation**

- 80+ checkpoints across all phases
- Pre-implementation planning
- Directory setup
- Configuration implementation
- STT implementation details
- TTS implementation details
- Testing requirements
- Security review
- Performance considerations
- Final verification
- Go-live checklist

### 5. This File (INDEX.md)

**Navigation and overview of all documentation**

---

## 🎯 Quick Navigation

### I want to...

**Add a new speech-to-text provider**

1. Start with [QUICKSTART.md](QUICKSTART.md) to understand the basics
2. Follow [DEVELOPMENT.md](DEVELOPMENT.md) step-by-step
3. Refer to [ARCHITECTURE.md](ARCHITECTURE.md) for design patterns
4. Use [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md) to verify completeness

**Add a new text-to-speech provider**

1. Same as above, but focus on the TTS section
2. Review [google/tts.go](google/tts.go) or [azure/tts.go](azure/tts.go) as examples

**Understand the architecture**

1. Read [ARCHITECTURE.md](ARCHITECTURE.md) for diagrams and patterns
2. Review sequence diagrams for data flow
3. Check component structure and concurrency model

**Debug a transformer issue**

1. Check [DEVELOPMENT.md](DEVELOPMENT.md) Troubleshooting section
2. Review error handling patterns in [ARCHITECTURE.md](ARCHITECTURE.md)
3. Look at similar provider's implementation for comparison

**Implement proper error handling**

1. See Error Handling Flow in [ARCHITECTURE.md](ARCHITECTURE.md)
2. Review best practices in [DEVELOPMENT.md](DEVELOPMENT.md)
3. Check existing providers for patterns

---

## 📖 Learning Path

### For Complete Beginners

1. **QUICKSTART.md** → Overview and 5-step summary
2. **ARCHITECTURE.md** → Visual diagrams and flows
3. **DEVELOPMENT.md** → Detailed implementation guide
4. **Existing Code** → Review google/, azure/, deepgram/ implementations
5. **PROVIDER_CHECKLIST.md** → Verify your implementation

### For Experienced Go Developers

1. **QUICKSTART.md** → 5-minute overview
2. **ARCHITECTURE.md** → Review design patterns
3. **DEVELOPMENT.md** → Reference as needed
4. **PROVIDER_CHECKLIST.md** → Pre-submission verification

### For Code Reviewers

1. **PROVIDER_CHECKLIST.md** → Review against checklist
2. **ARCHITECTURE.md** → Verify design patterns
3. **DEVELOPMENT.md** → Check best practices
4. **Existing Implementations** → Compare code style

---

## 📁 Package Structure

```
transformer/
├── INDEX.md                          ← You are here
├── QUICKSTART.md                     ← 5-minute overview
├── DEVELOPMENT.md                    ← Complete guide
├── ARCHITECTURE.md                   ← Diagrams & patterns
├── PROVIDER_CHECKLIST.md             ← 80+ checkpoints
├── README.md                         ← General info
├── transformer.go                    ← Core interfaces
│
├── google/                           ← Reference implementation
│   ├── google.go                     ← Configuration
│   ├── stt.go                        ← Speech-to-Text
│   ├── tts.go                        ← Text-to-Speech
│   └── README.md
│
├── azure/                            ← Event-driven example
│   ├── azure.go
│   ├── stt.go
│   ├── tts.go
│   └── README.md
│
├── deepgram/                         ← WebSocket example
│   ├── deepgram.go
│   ├── stt.go
│   ├── internal/
│   │   ├── stt_callback.go
│   │   └── type.go
│   └── README.md
│
├── assembly-ai/                      ← WebSocket with headers
│   ├── assemblyai.go
│   ├── stt.go
│   ├── internal/
│   │   └── type.go
│   └── README.md
│
└── [other providers...]
    ├── [provider].go
    ├── stt.go (if applicable)
    ├── tts.go (if applicable)
    └── README.md
```

---

## 🔑 Key Concepts

### Transformer Interface

Generic interface that defines the transformation pipeline:

- `Initialize()` - Setup connection and resources
- `Transform(ctx, input, options)` - Process input and deliver via callbacks
- `Close(ctx)` - Cleanup resources
- See [DEVELOPMENT.md](DEVELOPMENT.md) for details

### Speech-to-Text Transformer

Converts audio ([]byte) to transcribed text via `OnTranscript` callback:

- Receives audio data in chunks
- Returns transcript with confidence and language
- Indicates completion with `isCompleted` flag
- Example: [google/stt.go](google/stt.go)

### Text-to-Speech Transformer

Converts text (string) to audio ([]byte) via callbacks:

- Receives text to synthesize
- Returns audio chunks via `OnSpeech` callback
- Signals completion via `OnComplete` callback
- Example: [google/tts.go](google/tts.go)

### Callback Pattern

Results delivered through callbacks instead of return values:

- Enables streaming and real-time processing
- Allows multiple results from single Transform() call
- Supports error handling in callback execution
- See [ARCHITECTURE.md](ARCHITECTURE.md) for callback flow

### Thread Safety

All shared state protected by mutex:

- Provider client/connection
- Active streams
- Context tracking (contextId for TTS)
- Lock held only briefly
- See [ARCHITECTURE.md](ARCHITECTURE.md) concurrency model

---

## 🚀 Getting Started (TL;DR)

1. **Read**: [QUICKSTART.md](QUICKSTART.md) (5 minutes)
2. **Plan**: Gather provider API documentation and authentication method
3. **Create**: Directory structure following [QUICKSTART.md](QUICKSTART.md#5-step-structure-template)
4. **Implement**: STT or TTS following [DEVELOPMENT.md](DEVELOPMENT.md)
5. **Test**: Create unit and integration tests
6. **Review**: Check against [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)
7. **Reference**: Compare with similar provider (google, azure, deepgram)

---

## 📝 Code Examples

### Basic STT Provider Skeleton

```go
// 1. Create myprovider/myprovider.go
type myproviderOption struct {
    logger commons.Logger
    apiKey string
    audioConfig *protos.AudioConfig
    mdlOpts utils.Option
}

// 2. Create myprovider/stt.go
type myproviderSpeechToText struct {
    *myproviderOption
    mu sync.Mutex
    ctx context.Context
    ctxCancel context.CancelFunc
    logger commons.Logger
    client interface{}
    options *SpeechToTextInitializeOptions
}

func (m *myproviderSpeechToText) Initialize() error { }
func (m *myproviderSpeechToText) Transform(ctx, audio, opts) error { }
func (m *myproviderSpeechToText) Close(ctx) error { }
func (m *myproviderSpeechToText) Name() string { return "myprovider-speech-to-text" }
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for complete templates.

---

## 🧪 Testing Examples

```go
// Test new provider
func TestNewMyproviderOption(t *testing.T) {
    credential := &protos.VaultCredential{...}
    opts, err := NewMyproviderOption(logger, credential, audioConfig, modelOpts)
    require.NoError(t, err)
    require.NotNil(t, opts)
}

func TestTransform(t *testing.T) {
    transformer, _ := NewMyproviderSpeechToText(ctx, logger, credential, opts)
    transformer.Initialize()

    err := transformer.Transform(ctx, audioData, &SpeechToTextOption{})
    require.NoError(t, err)

    transformer.Close(ctx)
}
```

See [DEVELOPMENT.md](DEVELOPMENT.md) Testing section for complete examples.

---

## ✅ Pre-Submission Checklist

Before submitting your new provider:

- [ ] All files have copyright headers
- [ ] Passes `go test ./...` with no race conditions
- [ ] Follows naming conventions (see [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md))
- [ ] All exported functions have godoc comments
- [ ] No hardcoded credentials
- [ ] Proper error handling and logging
- [ ] Thread-safe implementation (mutex protection)
- [ ] Goroutine cleanup on Close()
- [ ] Callbacks handle nil safely
- [ ] Reviewed against [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)

---

## 📖 Reference Implementations

### Recommended Learning Path

| Complexity   | Provider     | Location       | Best For Learning             |
| ------------ | ------------ | -------------- | ----------------------------- |
| Beginner     | Google Cloud | [google/]      | Architecture & best practices |
| Beginner     | Azure        | [azure/]       | Event-driven patterns         |
| Intermediate | Deepgram     | [deepgram/]    | WebSocket implementation      |
| Intermediate | AssemblyAI   | [assembly-ai/] | WebSocket configuration       |
| Advanced     | Cartesia     | [cartesia/]    | Real-time streaming           |

### All Available Providers

#### Speech-to-Text (STT) Only

- **[AssemblyAI](assembly-ai/)** - WebSocket streaming, real-time transcription
- **[RevAI](revai/)** - High-accuracy transcription with rich metadata
- **[Speechmatics](speechmatics/)** - Multilingual support with advanced features

#### Text-to-Speech (TTS) Only

- **[ElevenLabs](elevenlabs/)** - High-quality voice synthesis with voice cloning
- **[Resemble](resemble/)** - Custom voice creation and synthesis

#### Full Support (STT + TTS)

- **[Google Cloud](google/)** ⭐ - Recommended for learning, comprehensive implementation
- **[Azure](azure/)** - Event-driven architecture, natural voice synthesis
- **[Deepgram](deepgram/)** - WebSocket streaming, real-time processing
- **[Cartesia](cartesia/)** - Real-time synthesis, custom voice models
- **[OpenAI](openai/)** - Integration with GPT models
- **[AWS](aws/)** - Amazon's Polly and Transcribe services
- **[Sarvam](sarvam/)** - Indian language support for both STT and TTS

---

## ❓ FAQ

### Q: What's the difference between STT and TTS?

**A:**

- **STT** (Speech-to-Text): Converts audio input to text output
- **TTS** (Text-to-Speech): Converts text input to audio output

### Q: Do I need to implement both STT and TTS?

**A:** No, implement only what the provider supports. Some only provide one.

### Q: How do callbacks work?

**A:** Results are delivered via callback functions rather than return values. This enables streaming and real-time results. See [ARCHITECTURE.md](ARCHITECTURE.md) Callback Delivery Pattern.

### Q: What if the provider doesn't support streaming?

**A:** You can batch/buffer the entire input and return all results at once via callback.

### Q: How do I handle provider-specific configuration?

**A:** Use `ModelOptions` (utils.Option) to access dynamic configuration. See [DEVELOPMENT.md](DEVELOPMENT.md) Configuration Handling section.

### Q: What about error handling?

**A:** Always log errors with provider prefix, wrap with context, and propagate to caller. See [ARCHITECTURE.md](ARCHITECTURE.md) Error Handling Flow.

---

## 📞 Support

1. **Questions about implementation**: See [DEVELOPMENT.md](DEVELOPMENT.md) Troubleshooting
2. **Need design review**: Check [ARCHITECTURE.md](ARCHITECTURE.md) patterns
3. **Missing checklist item**: Refer to [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)
4. **Need code example**: Search across google/, azure/, deepgram/ implementations
5. **Stuck on something**: Compare your code with similar provider in reference implementations

---

## 📋 Documentation Changelog

- **2025-01-11**: Initial documentation created
  - QUICKSTART.md - 5-minute overview
  - DEVELOPMENT.md - Comprehensive guide with templates
  - ARCHITECTURE.md - Diagrams and design patterns
  - PROVIDER_CHECKLIST.md - 80+ point verification checklist
  - INDEX.md - This file

---

**Happy implementing! 🎉**

Start with [QUICKSTART.md](QUICKSTART.md) for a 5-minute overview, then move to [DEVELOPMENT.md](DEVELOPMENT.md) for detailed guidance.
