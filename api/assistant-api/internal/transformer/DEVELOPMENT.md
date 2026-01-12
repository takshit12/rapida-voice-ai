# Adding a New Transformer - Development Guide

This guide explains how to add support for a new Speech-to-Text (STT) or Text-to-Speech (TTS) provider to the Rapida Voice AI platform.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Interfaces You Need to Implement](#interfaces-you-need-to-implement)
3. [Step-by-Step Guide](#step-by-step-guide)
4. [Implementation Example](#implementation-example)
5. [Testing Your Implementation](#testing-your-implementation)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

The transformer package provides a pluggable architecture for integrating different speech AI providers. The design follows these principles:

- **Interface-based**: All transformers implement standard interfaces
- **Provider-agnostic**: Core logic doesn't depend on specific implementations
- **Streaming support**: Built for real-time audio processing
- **Credential management**: Uses vault system for secure credential storage
- **Callback-driven**: Results delivered through callbacks, not polling

### Package Structure

```
transformer/
├── transformer.go          # Core interfaces
├── DEVELOPMENT.md          # This file
├── google/                 # Google provider
│   ├── google.go          # Configuration & setup
│   ├── stt.go             # Speech-to-text implementation
│   └── tts.go             # Text-to-speech implementation
├── deepgram/              # Deepgram provider
│   ├── deepgram.go
│   ├── stt.go
│   └── internal/          # Provider-specific types
├── [other-providers]/
└── README.md              # General documentation
```

---

## Interfaces You Need to Implement

### 1. Core Transformer Interface

All transformers implement the generic `Transformers` interface:

```go
type Transformers[IN any, opts TransformOption] interface {
    // Initialize sets up resources before transformation
    Initialize() error

    // Transform processes input and returns results via callbacks
    Transform(context.Context, IN, opts) error

    // Close cleans up resources
    Close(context.Context) error
}
```

### 2. Speech-to-Text Transformer

For STT providers, implement:

```go
type SpeechToTextTransformer interface {
    // Name returns the transformer identifier
    Name() string

    // Extends Transformers with []byte input (audio) and SpeechToTextOption
    Transformers[[]byte, *SpeechToTextOption]
}
```

**Callback Interface:**

```go
type SpeechToTextInitializeOptions struct {
    AudioConfig *protos.AudioConfig

    // Called when transcript is received
    OnTranscript func(
        transcript string,
        confidence float64,
        languages string,
        isCompleted bool,
    ) error

    // Model-specific options
    ModelOptions utils.Option
}
```

### 3. Text-to-Speech Transformer

For TTS providers, implement:

```go
type TextToSpeechTransformer interface {
    // Name returns the transformer identifier
    Name() string

    // Extends Transformers with string input (text) and TextToSpeechOption
    Transformers[string, *TextToSpeechOption]
}
```

**Callback Interface:**

```go
type TextToSpeechInitializeOptions struct {
    AudioConfig *protos.AudioConfig

    // Called when speech audio is generated
    OnSpeech func(string, []byte) error

    // Called when synthesis is complete
    OnComplete func(string) error

    // Model-specific options
    ModelOptions utils.Option
}
```

---

## Step-by-Step Guide

### Step 1: Create Provider Directory

Create a new directory under `transformer/` for your provider:

```bash
mkdir -p api/assistant-api/internal/transformer/myprovider
mkdir -p api/assistant-api/internal/transformer/myprovider/internal
```

### Step 2: Create Option Structure

Create `myprovider/myprovider.go` with configuration setup:

```go
package internal_transformer_myprovider

import (
    "fmt"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/pkg/utils"
    "github.com/rapidaai/protos"
)

// myproviderOption holds configuration for MyProvider
type myproviderOption struct {
    logger      commons.Logger
    apiKey      string
    audioConfig *protos.AudioConfig
    mdlOpts     utils.Option
}

// NewMyproviderOption initializes provider configuration
func NewMyproviderOption(
    logger commons.Logger,
    vaultCredential *protos.VaultCredential,
    audioConfig *protos.AudioConfig,
    opts utils.Option,
) (*myproviderOption, error) {
    // Extract credentials from vault
    credentialsMap := vaultCredential.GetValue().AsMap()

    apiKey, ok := credentialsMap["api_key"]
    if !ok {
        return nil, fmt.Errorf("myprovider: api_key not found in vault credentials")
    }

    return &myproviderOption{
        logger:      logger,
        apiKey:      apiKey.(string),
        audioConfig: audioConfig,
        mdlOpts:     opts,
    }, nil
}

// Helper method to convert audio format to provider-specific encoding
func (m *myproviderOption) GetEncoding() string {
    switch m.audioConfig.GetAudioFormat() {
    case protos.AudioConfig_LINEAR16:
        return "linear16"  // or provider's equivalent
    case protos.AudioConfig_MuLaw8:
        return "mulaw"
    default:
        return "linear16"
    }
}
```

### Step 3: Implement Speech-to-Text (STT)

Create `myprovider/stt.go`:

```go
package internal_transformer_myprovider

import (
    "context"
    "fmt"
    "sync"

    internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/protos"
)

type myproviderSpeechToText struct {
    *myproviderOption
    mu sync.Mutex

    // Context management
    ctx       context.Context
    ctxCancel context.CancelFunc

    logger  commons.Logger
    options *internal_transformer.SpeechToTextInitializeOptions

    // Provider-specific client (e.g., WebSocket, gRPC connection)
    client interface{}
}

// NewMyproviderSpeechToText creates a new STT transformer
func NewMyproviderSpeechToText(
    ctx context.Context,
    logger commons.Logger,
    credential *protos.VaultCredential,
    opts *internal_transformer.SpeechToTextInitializeOptions,
) (internal_transformer.SpeechToTextTransformer, error) {
    providerOpts, err := NewMyproviderOption(logger, credential, opts.AudioConfig, opts.ModelOptions)
    if err != nil {
        logger.Errorf("myprovider-stt: failed to initialize options: %v", err)
        return nil, err
    }

    ctx, cancel := context.WithCancel(ctx)
    return &myproviderSpeechToText{
        myproviderOption: providerOpts,
        ctx:              ctx,
        ctxCancel:        cancel,
        logger:           logger,
        options:          opts,
    }, nil
}

// Name returns the transformer identifier
func (m *myproviderSpeechToText) Name() string {
    return "myprovider-speech-to-text"
}

// Initialize establishes connection to the provider
func (m *myproviderSpeechToText) Initialize() error {
    // TODO: Establish connection (WebSocket, gRPC, REST, etc.)
    // Set up any necessary handlers/callbacks
    // Start listening goroutine if needed

    m.logger.Debugf("myprovider-stt: connection established")
    return nil
}

// Transform sends audio data to the provider for transcription
func (m *myproviderSpeechToText) Transform(
    ctx context.Context,
    audioData []byte,
    opts *internal_transformer.SpeechToTextOption,
) error {
    m.mu.Lock()
    client := m.client
    m.mu.Unlock()

    if client == nil {
        return fmt.Errorf("myprovider-stt: not initialized")
    }

    // TODO: Send audio data to provider
    // Handle any errors appropriately

    return nil
}

// Close cleans up resources
func (m *myproviderSpeechToText) Close(ctx context.Context) error {
    m.ctxCancel()

    m.mu.Lock()
    defer m.mu.Unlock()

    // TODO: Close connection, stop listening goroutines

    m.logger.Debugf("myprovider-stt: connection closed")
    return nil
}
```

### Step 4: Implement Text-to-Speech (TTS)

Create `myprovider/tts.go` following the same pattern:

```go
package internal_transformer_myprovider

import (
    "context"
    "fmt"
    "sync"

    internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/protos"
)

type myproviderTextToSpeech struct {
    *myproviderOption
    mu sync.Mutex

    ctx       context.Context
    ctxCancel context.CancelFunc

    logger  commons.Logger
    options *internal_transformer.TextToSpeechInitializeOptions
    client  interface{}
}

// NewMyproviderTextToSpeech creates a new TTS transformer
func NewMyproviderTextToSpeech(
    ctx context.Context,
    logger commons.Logger,
    credential *protos.VaultCredential,
    opts *internal_transformer.TextToSpeechInitializeOptions,
) (internal_transformer.TextToSpeechTransformer, error) {
    providerOpts, err := NewMyproviderOption(logger, credential, opts.AudioConfig, opts.ModelOptions)
    if err != nil {
        logger.Errorf("myprovider-tts: failed to initialize options: %v", err)
        return nil, err
    }

    ctx, cancel := context.WithCancel(ctx)
    return &myproviderTextToSpeech{
        myproviderOption: providerOpts,
        ctx:              ctx,
        ctxCancel:        cancel,
        logger:           logger,
        options:          opts,
    }, nil
}

// Name returns the transformer identifier
func (m *myproviderTextToSpeech) Name() string {
    return "myprovider-text-to-speech"
}

// Initialize establishes connection to the provider
func (m *myproviderTextToSpeech) Initialize() error {
    // TODO: Establish connection
    m.logger.Debugf("myprovider-tts: connection established")
    return nil
}

// Transform sends text to the provider for synthesis
func (m *myproviderTextToSpeech) Transform(
    ctx context.Context,
    text string,
    opts *internal_transformer.TextToSpeechOption,
) error {
    m.mu.Lock()
    client := m.client
    m.mu.Unlock()

    if client == nil {
        return fmt.Errorf("myprovider-tts: not initialized")
    }

    // TODO: Send text to provider for synthesis
    // Call m.options.OnSpeech(contextId, audioData) when audio arrives
    // Call m.options.OnComplete(contextId) when done

    return nil
}

// Close cleans up resources
func (m *myproviderTextToSpeech) Close(ctx context.Context) error {
    m.ctxCancel()

    m.mu.Lock()
    defer m.mu.Unlock()

    // TODO: Close connection

    m.logger.Debugf("myprovider-tts: connection closed")
    return nil
}
```

### Step 5: Create Internal Types (Optional)

For provider-specific types, create `myprovider/internal/type.go`:

```go
package myprovider_internal

// Define provider-specific message types, structures, etc.
```

### Step 6: Register the Transformer

Update the main transformer factory (if one exists) to include your new provider in the registration logic.

---

## Implementation Example

Here's a concrete example using Google Cloud Speech-to-Text as reference:

### Key Features to Consider:

1. **Connection Management**

   - Thread-safe client access with mutex
   - Proper connection lifecycle (Initialize → Transform → Close)

2. **Callback Handling**

   - Deliver results via callbacks, not return values
   - Handle callback errors appropriately
   - Ensure callbacks are nil-safe if optional

3. **Error Handling**

   - Log all errors with context
   - Wrap errors with `fmt.Errorf("%w", err)`
   - Propagate critical errors to caller

4. **Context Management**

   - Create child context with `context.WithCancel()`
   - Monitor context cancellation in listening goroutines
   - Clean up goroutines on cancellation

5. **Concurrency**
   - Use mutex for shared state
   - Protect client access
   - Ensure goroutine safety

---

## Testing Your Implementation

### Unit Tests

Create `myprovider/stt_test.go`:

```go
package internal_transformer_myprovider

import (
    "context"
    "testing"

    internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/protos"
)

func TestNewMyproviderSpeechToText(t *testing.T) {
    logger := commons.NewMockLogger() // or your test logger

    credential := &protos.VaultCredential{
        Value: &protos.VaultCredential_Map{
            Map: &protos.MapValue{
                Value: map[string]*protos.Value{
                    "api_key": {Value: &protos.Value_String{String: "test-key"}},
                },
            },
        },
    }

    opts := &internal_transformer.SpeechToTextInitializeOptions{
        AudioConfig: &protos.AudioConfig{
            AudioFormat: protos.AudioConfig_LINEAR16,
            SampleRate:  16000,
        },
        OnTranscript: func(transcript string, confidence float64, language string, isCompleted bool) error {
            t.Logf("Transcript: %s (confidence: %.2f)", transcript, confidence)
            return nil
        },
    }

    transformer, err := NewMyproviderSpeechToText(context.Background(), logger, credential, opts)
    if err != nil {
        t.Fatalf("Failed to create transformer: %v", err)
    }

    if transformer.Name() != "myprovider-speech-to-text" {
        t.Errorf("Expected name 'myprovider-speech-to-text', got '%s'", transformer.Name())
    }
}

func TestInitialize(t *testing.T) {
    // Test Initialize() method
    // Verify connection is established
}

func TestTransform(t *testing.T) {
    // Test Transform() with sample audio data
    // Verify callbacks are called correctly
}
```

### Integration Tests

Test with actual provider API (if available):

```go
func TestIntegrationWithRealProvider(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Use real credentials from environment
    // Test actual speech-to-text functionality
}
```

---

## Best Practices

### 1. **Credential Management**

```go
// ✅ GOOD: Extract and validate all required credentials
credMap := vaultCredential.GetValue().AsMap()
apiKey, ok := credMap["api_key"]
if !ok {
    return nil, fmt.Errorf("api_key not found in vault")
}

// ❌ BAD: Hardcoded credentials
client := NewClient(option.WithAPIKey("YOUR_API_KEY"))
```

### 2. **Error Handling**

```go
// ✅ GOOD: Log and wrap errors
if err != nil {
    m.logger.Errorf("myprovider-stt: operation failed: %v", err)
    return fmt.Errorf("operation failed: %w", err)
}

// ❌ BAD: Silent failures
if err != nil {
    // Ignore error
}
```

### 3. **Thread Safety**

```go
// ✅ GOOD: Protect shared state
m.mu.Lock()
client := m.client
m.mu.Unlock()

// ❌ BAD: Direct access without protection
return m.client.Send(data)
```

### 4. **Resource Cleanup**

```go
// ✅ GOOD: Clean up in Close()
func (m *myprovider) Close(ctx context.Context) error {
    m.ctxCancel()
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.client != nil {
        m.client.Close()
    }
    return nil
}

// ❌ BAD: Leaked resources
func (m *myprovider) Close(ctx context.Context) error {
    return nil
}
```

### 5. **Logging Standards**

```go
// ✅ GOOD: Consistent prefix with provider name
m.logger.Debugf("myprovider-stt: connection established")
m.logger.Errorf("myprovider-stt: failed to send audio: %v", err)

// ❌ BAD: No context
m.logger.Debugf("connection established")
```

### 6. **Configuration Handling**

```go
// ✅ GOOD: Validate and provide defaults
if language, err := m.mdlOpts.GetString("listen.language"); err == nil {
    opts.Language = language
} else {
    opts.Language = defaultLanguage
    m.logger.Debugf("Using default language: %s", defaultLanguage)
}
```

### 7. **Callback Safety**

```go
// ✅ GOOD: Nil check before calling optional callbacks
if m.options != nil && m.options.OnTranscript != nil {
    err := m.options.OnTranscript(transcript, confidence, language, true)
    if err != nil {
        m.logger.Errorf("callback error: %v", err)
    }
}
```

---

## Troubleshooting

### Common Issues

#### 1. "Provider not initialized" Error

**Cause:** Transform called before Initialize  
**Solution:** Ensure Initialize() returns successfully before calling Transform()

#### 2. Credentials Not Found

**Cause:** Wrong key names in vault  
**Solution:** Verify vault credential structure matches your implementation

#### 3. Callback Not Called

**Cause:** Listening goroutine crashed or callback is nil  
**Solution:**

- Add nil checks for callbacks
- Log all goroutine errors
- Verify error handling in message processing

#### 4. Memory Leaks

**Cause:** Goroutines not cleaned up on Close()  
**Solution:**

- Always cancel context in Close()
- Wait for goroutines to exit
- Verify no buffered channels without readers

#### 5. Race Conditions

**Cause:** Unsynchronized access to shared fields  
**Solution:**

- Use mutex for all shared state
- Always lock before accessing m.client, m.stream, etc.

---

## Checklist for New Transformer

- [ ] Created provider directory structure
- [ ] Implemented myprovider.go with option struct
- [ ] Implemented STT transformer (stt.go)
- [ ] Implemented TTS transformer (tts.go)
- [ ] Added proper error handling and logging
- [ ] Used mutex for thread-safe access
- [ ] Implemented context cancellation
- [ ] Added callback nil checks
- [ ] Created unit tests
- [ ] Tested with real provider (if possible)
- [ ] Documented any provider-specific configuration
- [ ] Followed naming conventions
- [ ] Added copyright header to all files
- [ ] Registered provider in factory/registry

---

## Available Provider Implementations

This section provides quick reference to all available providers in the transformer package.

### Speech-to-Text (STT) Only Providers

#### AssemblyAI

- **Location**: [assembly-ai/](assembly-ai/)
- **Type**: WebSocket-based streaming
- **Key Features**: Real-time transcription, custom language models
- **Configuration**: Requires API key, supports WebSocket headers
- **Best For**: Learning WebSocket implementation patterns
- **Files**:
  - `assemblyai.go` - Configuration and client setup
  - `stt.go` - STT implementation
  - `internal/type.go` - Provider-specific types

#### RevAI

- **Location**: [revai/](revai/)
- **Type**: REST API with polling
- **Key Features**: High-accuracy transcription, speaker identification
- **Configuration**: API token-based authentication
- **Best For**: REST API patterns
- **Files**:
  - `stt.go` - STT implementation with polling

#### Speechmatics

- **Location**: [speechmatics/](speechmatics/)
- **Type**: WebSocket-based streaming
- **Key Features**: Multilingual support, entity recognition
- **Configuration**: API token, region-based endpoints
- **Best For**: Multilingual STT implementations
- **Files**:
  - `stt.go` - STT implementation

### Text-to-Speech (TTS) Only Providers

#### ElevenLabs

- **Location**: [elevenlabs/](elevenlabs/)
- **Type**: REST API with streaming response
- **Key Features**: High-quality voices, voice cloning, emotion control
- **Configuration**: API key with model selection
- **Best For**: High-quality voice synthesis
- **Files**:
  - `elevenlabs.go` - Configuration and setup
  - `tts.go` - TTS implementation
  - `internal/type.go` - Provider-specific types

#### Resemble

- **Location**: [resemble/](resemble/)
- **Type**: REST API
- **Key Features**: Voice cloning, custom voice creation
- **Configuration**: API token, project ID
- **Best For**: Custom voice cloning implementations
- **Files**:
  - `resemble.go` - Configuration and setup
  - `tts.go` - TTS implementation
  - `internal/type.go` - Provider-specific types

### Multi-Purpose Providers (STT + TTS)

#### Google Cloud (Recommended for Learning) ⭐

- **Location**: [google/](google/)
- **Type**: gRPC and REST APIs
- **Key Features**: Well-structured code, comprehensive error handling, streaming support
- **Configuration**: Service account JSON, project ID
- **STT Capabilities**: Multiple languages, confidence scores, streaming
- **TTS Capabilities**: Multiple voices, SSML support, streaming audio
- **Best For**: Architecture reference, learning best practices
- **Files**:
  - `google.go` - Configuration with service account handling
  - `stt.go` - STT implementation with streaming
  - `tts.go` - TTS implementation with audio chunks
  - `internal/type.go` - gRPC request/response types
- **Learning Path**: Start here if new to transformers

#### Azure Cognitive Services

- **Location**: [azure/](azure/)
- **Type**: REST API with WebSocket events
- **Key Features**: Event-driven architecture, natural voice synthesis
- **Configuration**: Subscription key, region, endpoint
- **STT Capabilities**: Multiple languages, confidence scores
- **TTS Capabilities**: Neural voices, SSML support
- **Best For**: Event-driven pattern implementations
- **Files**:
  - `azure.go` - Configuration with Azure SDK
  - `stt.go` - STT implementation
  - `tts.go` - TTS implementation
  - `internal/type.go` - Azure-specific types

#### Deepgram

- **Location**: [deepgram/](deepgram/)
- **Type**: WebSocket-based streaming
- **Key Features**: Real-time transcription, pre-recorded audio support, multiple models
- **Configuration**: API key, model selection
- **STT Capabilities**: Multiple languages, live transcription, diarization
- **TTS Capabilities**: Multiple voices, streaming synthesis
- **Best For**: WebSocket implementation patterns
- **Files**:
  - `deepgram.go` - Configuration and WebSocket setup
  - `stt.go` - STT WebSocket implementation
  - `tts.go` - TTS implementation
  - `internal/stt_callback.go` - Callback pattern examples
  - `internal/type.go` - Provider-specific types

#### Cartesia

- **Location**: [cartesia/](cartesia/)
- **Type**: WebSocket-based real-time
- **Key Features**: Real-time bidirectional communication, custom voice models
- **Configuration**: API key, voice configuration
- **STT Capabilities**: Real-time transcription
- **TTS Capabilities**: Real-time synthesis with voice customization
- **Best For**: Real-time bidirectional streaming
- **Files**:
  - `cartesia.go` - Configuration and WebSocket setup
  - `stt.go` - STT implementation
  - `tts.go` - TTS implementation
  - `internal/type.go` - Provider-specific types

#### OpenAI

- **Location**: [openai/](openai/)
- **Type**: REST API
- **Key Features**: GPT model integration, multi-format support
- **Configuration**: API key, model selection
- **STT Capabilities**: Multiple audio formats, language detection
- **TTS Capabilities**: Multiple voices, audio format options
- **Best For**: OpenAI/GPT integration
- **Files**:
  - `stt.go` - STT implementation
  - `tts.go` - TTS implementation

#### AWS (Amazon Polly & Transcribe)

- **Location**: [aws/](aws/)
- **Type**: AWS SDK REST APIs
- **Key Features**: AWS ecosystem integration, scalability
- **Configuration**: AWS credentials, region, service endpoints
- **STT Capabilities**: Transcribe service integration
- **TTS Capabilities**: Polly service integration with voice options
- **Best For**: AWS infrastructure integration
- **Files**:
  - `stt.go` - Transcribe implementation
  - `tts.go` - Polly implementation

#### Sarvam (Indian Language Support)

- **Location**: [sarvam/](sarvam/)
- **Type**: REST API with regional support
- **Key Features**: Indian language support (Hindi, Tamil, Telugu, etc.)
- **Configuration**: API key, language selection
- **STT Capabilities**: Multiple Indian languages
- **TTS Capabilities**: Multiple Indian language voices
- **Best For**: Regional language support
- **Files**:
  - `sarvam.go` - Configuration with language handling
  - `stt.go` - STT implementation
  - `tts.go` - TTS implementation
  - `internal/type.go` - Provider-specific types

### Provider Selection Guide

| Requirement         | Recommended Provider | Alternative  |
| ------------------- | -------------------- | ------------ |
| Learning/Reference  | Google Cloud         | Azure        |
| Real-time Streaming | Cartesia             | Deepgram     |
| High-Quality TTS    | ElevenLabs           | Azure        |
| Cost-Effective STT  | AssemblyAI           | Deepgram     |
| Voice Cloning       | Resemble             | ElevenLabs   |
| Regional Languages  | Sarvam               | Google Cloud |
| AWS Ecosystem       | AWS                  | -            |
| Multilingual        | Speechmatics         | Google Cloud |
| WebSocket Pattern   | Deepgram             | Cartesia     |
| Event-Driven        | Azure                | -            |

---

## References

- [Transformer Interface](transformer.go)
- [Provider Checklist](PROVIDER_CHECKLIST.md)
- [Quick Reference](QUICKSTART.md)
- [Architecture Details](ARCHITECTURE.md)

For questions or clarifications, refer to existing implementations as examples.
