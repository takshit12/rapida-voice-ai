# Transformer Package - Speech & Text Processing

The Transformer package provides a pluggable architecture for integrating multiple speech-to-text (STT) and text-to-speech (TTS) providers into the Rapida Voice AI platform.

## 🎯 What is a Transformer?

A Transformer is a provider-agnostic adapter that:

- Converts audio streams into transcribed text (STT)
- Converts text into audio streams (TTS)
- Manages connection lifecycle and error handling
- Delivers results through callbacks for real-time processing
- Handles resource cleanup and graceful shutdown

## 🏗️ Supported Providers

| Provider                     | STT | TTS | Status        | Notes                    |
| ---------------------------- | --- | --- | ------------- | ------------------------ |
| **Google Cloud**             | ✅  | ✅  | Production    | Recommended for learning |
| **Azure Cognitive Services** | ✅  | ✅  | Production    | Event-driven             |
| **Deepgram**                 | ✅  | ❌  | Production    | WebSocket streaming      |
| **AssemblyAI**               | ✅  | ❌  | Production    | WebSocket with headers   |
| **ElevenLabs**               | ❌  | ✅  | Production    | TTS only                 |
| **Cartesia**                 | ❌  | ✅  | Development   | Experimental             |
| **Sarvam**                   | ✅  | ✅  | Development   | Limited support          |
| **AWS**                      | ⚠️  | ⚠️  | Unimplemented | Placeholder only         |

## 📚 Documentation

Start here for adding a new provider:

| Document                                           | Time   | Purpose                     |
| -------------------------------------------------- | ------ | --------------------------- |
| **[INDEX.md](INDEX.md)**                           | 5 min  | Overview & navigation       |
| **[QUICKSTART.md](QUICKSTART.md)**                 | 5 min  | 5-step quick reference      |
| **[ARCHITECTURE.md](ARCHITECTURE.md)**             | 20 min | Diagrams & design patterns  |
| **[DEVELOPMENT.md](DEVELOPMENT.md)**               | 60 min | Step-by-step implementation |
| **[PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)** | Review | 80+ verification points     |

**Total Documentation**: 2,122 lines of comprehensive guides with diagrams, code examples, and checklists.

## 🚀 Quick Start: Add a New Provider (5 Steps)

```bash
# 1. Create directory
mkdir -p api/assistant-api/internal/transformer/myprovider/internal

# 2-5. Follow DEVELOPMENT.md step-by-step
```

See [QUICKSTART.md](QUICKSTART.md) for the complete 5-step guide.

## 📖 Learning Paths

### First Time? (2 hours)

1. Read [INDEX.md](INDEX.md) (10 min)
2. Read [QUICKSTART.md](QUICKSTART.md) (10 min)
3. Read [ARCHITECTURE.md](ARCHITECTURE.md) (20 min)
4. Follow [DEVELOPMENT.md](DEVELOPMENT.md) (60 min)
5. Use [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md) (verification)

### Experienced Go Dev? (30-45 min)

1. Skim [QUICKSTART.md](QUICKSTART.md)
2. Review [ARCHITECTURE.md](ARCHITECTURE.md) patterns
3. Use [DEVELOPMENT.md](DEVELOPMENT.md) templates
4. Check [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)

### Just Want Reference Implementations?

- [google/](google/) - Comprehensive, well-structured
- [deepgram/](deepgram/) - WebSocket streaming
- [azure/](azure/) - Event-driven callbacks
- [assembly-ai/](assembly-ai/) - Custom WebSocket headers

## 🔧 Core Concepts

### Transformer Interface

```go
// All transformers implement this generic interface
type Transformers[IN any, opts TransformOption] interface {
    Initialize() error                          // Setup connection
    Transform(context.Context, IN, opts) error  // Process input
    Close(context.Context) error                // Cleanup
}
```

### Speech-to-Text (STT)

```
Input: []byte (audio data)
Output: Callback with OnTranscript(text, confidence, language, isFinal)
```

### Text-to-Speech (TTS)

```
Input: string (text to synthesize)
Output: Callbacks with OnSpeech(contextId, audioData) and OnComplete(contextId)
```

## 📋 Directory Structure

```
transformer/
├── INDEX.md                          # Navigation hub
├── QUICKSTART.md                     # 5-minute overview
├── DEVELOPMENT.md                    # Complete implementation guide
├── ARCHITECTURE.md                   # Diagrams & patterns
├── PROVIDER_CHECKLIST.md             # 80+ checkpoints
├── DOCUMENTATION_SUMMARY.md          # This documentation summary
├── transformer.go                    # Core interfaces
│
├── google/                           # Google Cloud (STT + TTS)
│   ├── google.go                     # Configuration
│   ├── stt.go                        # Speech-to-Text
│   ├── tts.go                        # Text-to-Speech
│   └── README.md                     # Provider-specific docs
│
├── azure/                            # Azure (STT + TTS)
├── deepgram/                         # Deepgram (STT)
│   ├── internal/
│   │   ├── stt_callback.go           # Callback handler
│   │   └── type.go                   # Message types
│   └── ...
│
├── assembly-ai/                      # AssemblyAI (STT)
├── elevenlabs/                       # ElevenLabs (TTS)
├── cartesia/                         # Cartesia (TTS)
├── sarvam/                           # Sarvam (STT + TTS)
└── revai/                            # RevAI (stub)
```

## 🎯 Implementation Checklist

- [ ] Create provider directory
- [ ] Create configuration (myprovider.go)
- [ ] Implement STT transformer (stt.go)
- [ ] Implement TTS transformer (tts.go)
- [ ] Create unit tests
- [ ] Add integration tests
- [ ] Verify thread safety
- [ ] Handle errors properly
- [ ] Clean up resources on Close()
- [ ] Verify against [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)

See [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md) for comprehensive 80-point checklist.

## 🔐 Security Notes

- ✅ All credentials loaded from vault (never hardcoded)
- ✅ Validate all required credentials in constructor
- ✅ Clear credentials on Close()
- ✅ No credentials in logs
- ✅ Use HTTPS/TLS for all connections
- ❌ Do NOT commit actual API keys

## 🧪 Testing

Each provider should have:

- Unit tests for initialization
- Unit tests for Transform() method
- Unit tests for callback handling
- Tests for resource cleanup
- Concurrency tests (run with `go test -race`)

Example from [DEVELOPMENT.md](DEVELOPMENT.md):

```go
func TestNewMyproviderOption(t *testing.T) {
    credential := &protos.VaultCredential{...}
    opts, err := NewMyproviderOption(logger, credential, audioConfig, modelOpts)
    require.NoError(t, err)
}
```

## 🚨 Common Issues

### "Provider not initialized"

→ Call `Initialize()` before `Transform()`

### Callbacks not triggered

→ Check if callback is nil, verify listening goroutine is running

### Memory leaks

→ Ensure `Close()` cancels context and stops goroutines

### Race conditions

→ Run `go test -race ./...` to detect

See [DEVELOPMENT.md](DEVELOPMENT.md) Troubleshooting for more.

## 📊 Code Statistics

| Aspect                    | Count |
| ------------------------- | ----- |
| Documentation Files       | 6     |
| Documentation Lines       | 2,122 |
| Code Diagrams             | 8+    |
| Code Examples             | 10+   |
| Supported Providers       | 8     |
| Reference Implementations | 4     |
| Checklist Points          | 80+   |

## 🎓 Reference Implementations

Learn from existing providers:

### Google Cloud (Recommended)

- **Use for**: Architecture reference, comprehensive example
- **Features**: Well-documented, error handling, streaming
- **Files**: [google/](google/)

### Deepgram (WebSocket Example)

- **Use for**: WebSocket implementation patterns
- **Features**: Real-time streaming, callback patterns
- **Files**: [deepgram/](deepgram/)

### Azure (Event-driven)

- **Use for**: Event-based callback patterns
- **Features**: Session management, event lifecycle
- **Files**: [azure/](azure/)

### AssemblyAI (Advanced WebSocket)

- **Use for**: Custom headers and authentication
- **Features**: WebSocket with query parameters
- **Files**: [assembly-ai/](assembly-ai/)

## 🔄 Lifecycle

```
┌─────────┐
│ Created │ ← NewMyprovider...()
└────┬────┘
     │
     ▼
┌──────────────┐
│ Initialize() │ ← Setup connection & resources
└────┬────────┘
     │
     ▼
┌────────────────┐
│ Transform()×N  │ ← Process audio/text, deliver via callbacks
└────┬───────────┘
     │
     ▼
┌─────────┐
│ Close() │ ← Cleanup resources
└─────────┘
```

## 📞 Support

1. **Questions?** Check [INDEX.md](INDEX.md) FAQ
2. **Stuck?** See [DEVELOPMENT.md](DEVELOPMENT.md) Troubleshooting
3. **Need patterns?** Review [ARCHITECTURE.md](ARCHITECTURE.md)
4. **Implementation help?** Follow [DEVELOPMENT.md](DEVELOPMENT.md) step-by-step
5. **Verification?** Use [PROVIDER_CHECKLIST.md](PROVIDER_CHECKLIST.md)

## 🎉 Getting Started

**Choose your path:**

- 🚀 **Quick Overview**: Start with [QUICKSTART.md](QUICKSTART.md) (5 min)
- 📚 **Complete Guide**: Start with [DEVELOPMENT.md](DEVELOPMENT.md) (60 min)
- 🏗️ **Understand Design**: Start with [ARCHITECTURE.md](ARCHITECTURE.md) (20 min)
- 🗺️ **Navigation**: Start with [INDEX.md](INDEX.md) (5 min)

---

## 📋 Provider Credentials Format

All providers use vault for credentials:

```json
{
  "key": "api-key-value",
  "project_id": "project-identifier",
  "subscription_key": "subscription-key",
  "endpoint": "https://api.example.com",
  "service_account_key": "json-string"
}
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for provider-specific credential requirements.

---

**Last Updated**: 2025-01-11  
**Documentation Version**: 1.0  
**Total Documentation**: 2,122 lines across 6 files

Start with [INDEX.md](INDEX.md) or [QUICKSTART.md](QUICKSTART.md) 👈
