# Telephony Package - Integration Implementation Guide

The telephony package provides a unified abstraction for integrating with multiple voice telephony providers (Twilio, Vonage, Exotel, etc.). It enables seamless phone call routing, WebSocket media streaming, and event handling while maintaining a consistent interface.

## Overview

The package currently supports the following telephony providers:

- **Twilio** - Market-leading voice platform with TwiML support
- **Vonage (Nexmo)** - Enterprise voice provider with NCCO support
- **Exotel** - Voice and SMS provider with HTTP APIs
- **[Your Provider]** - Ready for new integrations

---

## Architecture

### Factory Pattern

The package uses a factory pattern to instantiate telephony providers:

```go
import (
    telephony "github.com/rapidaai/api/assistant-api/internal/telephony"
)

// Get telephony provider
provider, err := telephony.GetTelephony(
    telephony.Twilio,
    assistantConfig,
    logger,
)
```

### Core Interface

All telephony implementations must satisfy the `Telephony` interface:

```go
type Telephony interface {
    // WebSocket streamer for real-time audio handling
    Streamer(
        ctx *gin.Context,
        connection *websocket.Conn,
        assistantID uint64,
        assistantVersion string,
        assistantConversationID uint64,
    ) internal_streamers.Streamer

    // Initiate an outbound call
    MakeCall(
        auth types.SimplePrinciple,
        toPhone string,
        fromPhone string,
        assistantId, assistantConversationId uint64,
        vaultCredential *protos.VaultCredential,
        opts utils.Option,
    ) ([]*types.Metadata, []*types.Metric, []*types.Event, error)

    // Handle call status events
    StatusCallback(
        ctx *gin.Context,
        auth types.SimplePrinciple,
        assistantId, assistantConversationId uint64,
    ) ([]*types.Metric, []*types.Event, error)

    // Catch-all event handler
    CatchAllStatusCallback(ctx *gin.Context) (*string, []*types.Metric, []*types.Event, error)

    // Answer incoming calls with TwiML/NCCO
    IncomingCall(
        c *gin.Context,
        auth types.SimplePrinciple,
        assistantId uint64,
        clientNumber string,
        assistantConversationId uint64,
    ) error

    // Extract call details from incoming request
    AcceptCall(c *gin.Context) (client *string, assistantId *string, err error)
}
```

### Streamer Interface

The WebSocket streamer handles real-time audio:

```go
type Streamer interface {
    // Get the context for the stream session
    Context() context.Context

    // Receive audio/messages from the provider
    Recv() (*protos.AssistantMessagingRequest, error)

    // Send audio/responses to the provider
    Send(response *protos.AssistantMessagingResponse) error

    // Clean up resources
    Close() error
}
```

---

## Call Flow

```
1. Incoming Call
   ↓
   [Telephony Provider] → [AcceptCall()] → Extract client info
   ↓
2. Answer Call
   ↓
   [IncomingCall()] → Return TwiML/NCCO with WebSocket URL
   ↓
3. WebSocket Connection
   ↓
   [Streamer()] → Create media handler
   ↓
4. Real-time Audio Exchange
   ↓
   [Recv()] → Audio from caller
   [Send()] → Audio to caller
   ↓
5. Call Events
   ↓
   [StatusCallback()] → Handle status updates (ringing, answered, completed)
   ↓
6. Call Termination
   ↓
   [Streamer.Close()] → Clean up WebSocket
```

---

## Step-by-Step Implementation Guide

### Step 1: Create Provider Directory

```
telephony/
├── my-provider/
│   ├── my_provider.go              # Main telephony implementation
│   ├── my_provider_websocket.go    # WebSocket streamer
│   └── internal/
│       └── (optional) audio processing utilities
```

### Step 2: Implement Main Provider (my_provider.go)

```go
package internal_myprovider_telephony

import (
    "fmt"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/rapidaai/api/assistant-api/config"
    internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
    internal_type "github.com/rapidaai/api/assistant-api/internal/type"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/pkg/types"
    "github.com/rapidaai/pkg/utils"
    "github.com/rapidaai/protos"
)

type myProviderTelephony struct {
    appCfg *config.AssistantConfig
    logger commons.Logger
}

// Factory function - called by the main factory
func NewMyProviderTelephony(
    config *config.AssistantConfig,
    logger commons.Logger,
) (internal_type.Telephony, error) {
    return &myProviderTelephony{
        appCfg: config,
        logger: logger,
    }, nil
}

// ============================================
// 1. ACCEPT INCOMING CALL
// ============================================

// AcceptCall extracts caller information from the incoming request
func (mpt *myProviderTelephony) AcceptCall(c *gin.Context) (*string, *string, error) {
    // Extract query parameters or parse request body
    // to get caller number and assistant ID

    // Example for Twilio:
    // FromNumber := c.PostForm("From")
    // AssistantId := c.PostForm("AssistantId")

    // Example for Vonage:
    // var payload map[string]interface{}
    // json.NewDecoder(c.Request.Body).Decode(&payload)
    // FromNumber := payload["from"].(string)

    mpt.logger.Debugf("my-provider: accepting call from: %s", "clientNumber")

    clientNumber := "1234567890"  // TODO: Extract from request
    assistantId := "12345"         // TODO: Extract from request

    return &clientNumber, &assistantId, nil
}

// ============================================
// 2. HANDLE INCOMING CALL
// ============================================

// IncomingCall responds with connection instructions (TwiML/NCCO/etc)
func (mpt *myProviderTelephony) IncomingCall(
    c *gin.Context,
    auth types.SimplePrinciple,
    assistantId uint64,
    clientNumber string,
    assistantConversationId uint64,
) error {
    // Create URL for WebSocket media stream
    wsUrl := fmt.Sprintf(
        "wss://%s/%s",
        mpt.appCfg.PublicAssistantHost,
        internal_type.GetAnswerPath("my-provider", auth, assistantId, assistantConversationId, clientNumber),
    )

    // Create URL for status callbacks
    statusUrl := fmt.Sprintf(
        "https://%s/%s",
        mpt.appCfg.PublicAssistantHost,
        internal_type.GetEventPath("my-provider", auth, assistantId, assistantConversationId),
    )

    // Build provider-specific connection instruction
    // This tells the provider where to send audio and events

    // TODO: For Twilio, this would be TwiML
    // TODO: For Vonage, this would be NCCO
    // TODO: For Exotel, this would be HTTP JSON

    // Example response structure:
    instructions := map[string]interface{}{
        "action": "connect",
        "endpoint": map[string]string{
            "type": "websocket",
            "uri":  wsUrl,
        },
        "statusCallback": statusUrl,
    }

    c.JSON(200, instructions)
    mpt.logger.Debugf("my-provider: incoming call answered with WebSocket URL: %s", wsUrl)
    return nil
}

// ============================================
// 3. CREATE OUTBOUND CALL
// ============================================

// MakeCall initiates an outbound call
func (mpt *myProviderTelephony) MakeCall(
    auth types.SimplePrinciple,
    toPhone string,
    fromPhone string,
    assistantId, assistantConversationId uint64,
    vaultCredential *protos.VaultCredential,
    opts utils.Option,
) ([]*types.Metadata, []*types.Metric, []*types.Event, error) {
    // Initialize metadata and metrics
    mtds := []*types.Metadata{
        types.NewMetadata("telephony.toPhone", toPhone),
        types.NewMetadata("telephony.fromPhone", fromPhone),
        types.NewMetadata("telephony.provider", "my-provider"),
    }
    event := []*types.Event{
        types.NewEvent("api-call", map[string]interface{}{}),
    }

    // 1. Authenticate with provider
    client, err := mpt.createClient(vaultCredential, opts)
    if err != nil {
        mpt.logger.Errorf("my-provider: authentication failed: %v", err)
        mtds = append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("auth error: %s", err.Error())))
        return mtds, []*types.Metric{types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))}, event, err
    }

    // 2. Create WebSocket and status callback URLs
    wsUrl := fmt.Sprintf(
        "wss://%s/%s",
        mpt.appCfg.PublicAssistantHost,
        internal_type.GetAnswerPath("my-provider", auth, assistantId, assistantConversationId, toPhone),
    )
    statusUrl := fmt.Sprintf(
        "https://%s/%s",
        mpt.appCfg.PublicAssistantHost,
        internal_type.GetEventPath("my-provider", auth, assistantId, assistantConversationId),
    )

    // 3. Call provider API to initiate the call
    callResponse, err := mpt.callProvider(client, toPhone, fromPhone, wsUrl, statusUrl)
    if err != nil {
        mpt.logger.Errorf("my-provider: call creation failed: %v", err)
        mtds = append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("call failed: %s", err.Error())))
        return mtds, []*types.Metric{types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))}, event, err
    }

    // 4. Extract call ID from response for reference
    mtds = append(mtds, types.NewMetadata("telephony.conversation_reference", callResponse.CallId))
    event = append(event, types.NewEvent(callResponse.Status, callResponse))

    mpt.logger.Infof("my-provider: call initiated successfully. CallID: %s", callResponse.CallId)
    return mtds, []*types.Metric{types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Status of telephony api"))}, event, nil
}

// ============================================
// 4. HANDLE STATUS CALLBACKS
// ============================================

// StatusCallback handles events from the provider
// Called when: call initiated, ringing, answered, completed, failed
func (mpt *myProviderTelephony) StatusCallback(
    c *gin.Context,
    auth types.SimplePrinciple,
    assistantId, assistantConversationId uint64,
) ([]*types.Metric, []*types.Event, error) {
    // Parse provider-specific status payload
    var statusPayload map[string]interface{}

    // TODO: Parse request based on provider format
    // For Twilio: URL-encoded form data
    // For Vonage: JSON body
    // For Exotel: Custom format

    mpt.logger.Infof("my-provider: status callback received: %+v", statusPayload)

    // Extract status
    status, ok := statusPayload["status"].(string)
    if !ok {
        status = "unknown"
    }

    // Create metrics and events
    metrics := []*types.Metric{
        types.NewMetric("STATUS", status, utils.Ptr("Call status")),
    }
    events := []*types.Event{
        types.NewEvent(status, statusPayload),
    }

    return metrics, events, nil
}

// CatchAllStatusCallback handles unexpected callbacks
func (mpt *myProviderTelephony) CatchAllStatusCallback(c *gin.Context) (*string, []*types.Metric, []*types.Event, error) {
    // Handle callbacks that don't match expected patterns
    mpt.logger.Warnf("my-provider: catch-all callback triggered")
    return nil, nil, nil, nil
}

// ============================================
// 5. CREATE WEBSOCKET STREAMER
// ============================================

// Streamer creates the media streaming handler
func (mpt *myProviderTelephony) Streamer(
    c *gin.Context,
    connection *websocket.Conn,
    assistantID uint64,
    assistantVersion string,
    assistantConversationID uint64,
) internal_streamers.Streamer {
    // Create and return WebSocket streamer
    // This will handle real-time audio exchange
    return NewMyProviderWebsocketStreamer(
        mpt.logger,
        connection,
        assistantID,
        assistantVersion,
        assistantConversationID,
    )
}

// ============================================
// PRIVATE HELPER METHODS
// ============================================

// createClient initializes provider client with credentials
func (mpt *myProviderTelephony) createClient(
    credential *protos.VaultCredential,
    opts utils.Option,
) (interface{}, error) {
    // Extract credentials from vault
    credMap := credential.GetValue().AsMap()

    apiKey, ok := credMap["api_key"]
    if !ok {
        return nil, fmt.Errorf("my-provider: api_key not found in credentials")
    }

    apiSecret, ok := credMap["api_secret"]
    if !ok {
        return nil, fmt.Errorf("my-provider: api_secret not found in credentials")
    }

    // TODO: Create and return provider's API client
    // Example: return myprovider.NewClient(apiKey, apiSecret), nil

    return nil, nil
}

// callProvider makes the actual API call to initiate a call
func (mpt *myProviderTelephony) callProvider(
    client interface{},
    toPhone, fromPhone, wsUrl, statusUrl string,
) (*CallResponse, error) {
    // TODO: Call provider's API to create the call
    // Steps:
    // 1. Format the request (phone numbers, URLs, etc.)
    // 2. Make HTTP request to provider API
    // 3. Parse response
    // 4. Handle errors
    // 5. Return call details (ID, status)

    return &CallResponse{
        CallId: "call-123",
        Status: "initiated",
    }, nil
}

// ============================================
// HELPER TYPES
// ============================================

type CallResponse struct {
    CallId string
    Status string
}
```

### Step 3: Implement WebSocket Streamer (my_provider_websocket.go)

```go
package internal_myprovider_telephony

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "sync"

    "github.com/gorilla/websocket"
    internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
    "github.com/rapidaai/pkg/commons"
    "github.com/rapidaai/protos"
)

type myProviderWebsocketStreamer struct {
    logger     commons.Logger
    conn       *websocket.Conn
    ctx        context.Context
    cancelFunc context.CancelFunc

    // Assistant info
    assistantId uint64
    version     string
    conversationId uint64

    // Audio buffers
    mu                 sync.Mutex
    inputAudioBuffer   []byte
    outputAudioBuffer  []byte

    // Provider-specific fields
    encoder *base64.Encoding
    streamId string
}

// Provider-specific event structure
type MyProviderMediaEvent struct {
    Event     string `json:"event"`
    StreamId  string `json:"stream_id"`
    Timestamp string `json:"timestamp"`
    Media     struct {
        Payload string `json:"payload"` // Base64-encoded audio
    } `json:"media"`
}

func NewMyProviderWebsocketStreamer(
    logger commons.Logger,
    connection *websocket.Conn,
    assistantId uint64,
    version string,
    conversationId uint64,
) internal_streamers.Streamer {
    ctx, cancel := context.WithCancel(context.Background())

    return &myProviderWebsocketStreamer{
        logger:     logger,
        conn:       connection,
        ctx:        ctx,
        cancelFunc: cancel,
        assistantId: assistantId,
        version:    version,
        conversationId: conversationId,
        encoder:    base64.StdEncoding,
    }
}

// Context returns the stream context
func (mws *myProviderWebsocketStreamer) Context() context.Context {
    return mws.ctx
}

// Recv receives audio from the caller
func (mws *myProviderWebsocketStreamer) Recv() (*protos.AssistantMessagingRequest, error) {
    if mws.conn == nil {
        return nil, fmt.Errorf("WebSocket connection is nil")
    }

    // Read message from WebSocket
    _, message, err := mws.conn.ReadMessage()
    if err != nil {
        if err == io.EOF {
            mws.logger.Infof("my-provider: WebSocket closed by peer")
        } else {
            mws.logger.Errorf("my-provider: failed to read message: %v", err)
        }
        return nil, err
    }

    // Parse provider-specific event format
    var event MyProviderMediaEvent
    if err := json.Unmarshal(message, &event); err != nil {
        mws.logger.Errorf("my-provider: failed to parse event: %v", err)
        return nil, fmt.Errorf("failed to parse event")
    }

    // Handle different event types
    switch event.Event {
    case "connected":
        mws.logger.Debugf("my-provider: stream connected, stream_id: %s", event.StreamId)
        mws.streamId = event.StreamId
        return nil, nil

    case "media":
        // Decode audio payload (provider-specific encoding)
        audioData, err := mws.decodeAudio(event.Media.Payload)
        if err != nil {
            mws.logger.Errorf("my-provider: failed to decode audio: %v", err)
            return nil, err
        }

        // Store audio in buffer
        mws.mu.Lock()
        mws.inputAudioBuffer = append(mws.inputAudioBuffer, audioData...)
        mws.mu.Unlock()

        // Create messaging request
        request := &protos.AssistantMessagingRequest{
            AssistantId:        mws.assistantId,
            AssistantConversationId: mws.conversationId,
            Audio: &protos.Audio{
                Data: audioData,
            },
        }

        mws.logger.Debugf("my-provider: received audio chunk, size: %d bytes", len(audioData))
        return request, nil

    case "closed":
        mws.logger.Infof("my-provider: stream closed")
        return nil, io.EOF

    default:
        mws.logger.Warnf("my-provider: unknown event type: %s", event.Event)
        return nil, nil
    }
}

// Send sends audio to the caller
func (mws *myProviderWebsocketStreamer) Send(response *protos.AssistantMessagingResponse) error {
    if mws.conn == nil {
        return fmt.Errorf("WebSocket connection is nil")
    }

    if response.Audio == nil {
        return fmt.Errorf("response has no audio data")
    }

    // Encode audio to provider format
    encodedAudio := mws.encodeAudio(response.Audio.Data)

    // Create provider-specific response structure
    mediaResponse := MyProviderMediaEvent{
        Event:    "media",
        StreamId: mws.streamId,
        Media: struct {
            Payload string `json:"payload"`
        }{
            Payload: encodedAudio,
        },
    }

    // Marshal to JSON
    payload, err := json.Marshal(mediaResponse)
    if err != nil {
        mws.logger.Errorf("my-provider: failed to marshal response: %v", err)
        return err
    }

    // Send through WebSocket
    mws.mu.Lock()
    defer mws.mu.Unlock()

    if err := mws.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
        mws.logger.Errorf("my-provider: failed to send audio: %v", err)
        return err
    }

    mws.mu.Lock()
    mws.outputAudioBuffer = append(mws.outputAudioBuffer, response.Audio.Data...)
    mws.mu.Unlock()

    mws.logger.Debugf("my-provider: sent audio chunk, size: %d bytes", len(response.Audio.Data))
    return nil
}

// Close closes the WebSocket connection
func (mws *myProviderWebsocketStreamer) Close() error {
    mws.logger.Debugf("my-provider: closing WebSocket streamer")

    mws.cancelFunc()

    mws.mu.Lock()
    defer mws.mu.Unlock()

    if mws.conn != nil {
        return mws.conn.Close()
    }

    return nil
}

// ============================================
// HELPER METHODS
// ============================================

// decodeAudio decodes provider-specific audio format
func (mws *myProviderWebsocketStreamer) decodeAudio(payload string) ([]byte, error) {
    // TODO: Implement provider-specific decoding
    // Common formats:
    // - Base64: base64.StdEncoding.DecodeString(payload)
    // - Hex: hex.DecodeString(payload)
    // - Raw: []byte(payload)

    audioData, err := base64.StdEncoding.DecodeString(payload)
    if err != nil {
        return nil, err
    }

    return audioData, nil
}

// encodeAudio encodes audio to provider-specific format
func (mws *myProviderWebsocketStreamer) encodeAudio(audioData []byte) string {
    // TODO: Implement provider-specific encoding
    // Common formats:
    // - Base64: base64.StdEncoding.EncodeToString(audioData)
    // - Hex: hex.EncodeToString(audioData)
    // - Raw: string(audioData)

    return base64.StdEncoding.EncodeToString(audioData)
}
```

### Step 4: Register Provider in Factory

Update [telephony.go](telephony.go):

```go
package internal_telephony_factory

import (
    "errors"

    "github.com/rapidaai/api/assistant-api/config"
    internal_myprovider_telephony "github.com/rapidaai/api/assistant-api/internal/telephony/my-provider"
    internal_type "github.com/rapidaai/api/assistant-api/internal/type"
    "github.com/rapidaai/pkg/commons"
)

type Telephony string

const (
    Twilio    Telephony = "twilio"
    Exotel    Telephony = "exotel"
    Vonage    Telephony = "vonage"
    MyProvider Telephony = "my-provider"  // ADD THIS
)

func GetTelephony(
    at Telephony,
    cfg *config.AssistantConfig,
    logger commons.Logger,
) (internal_type.Telephony, error) {
    switch at {
    case Twilio:
        return internal_twilio_telephony.NewTwilioTelephony(cfg, logger)
    case Exotel:
        return internal_exotel_telephony.NewExotelTelephony(cfg, logger)
    case Vonage:
        return internal_vonage_telephony.NewVonageTelephony(cfg, logger)
    case MyProvider:                        // ADD THIS
        return internal_myprovider_telephony.NewMyProviderTelephony(cfg, logger)
    default:
        return errors.New("illegal telephony provider")
    }
}
```

### Step 5: Test the Implementation

Create `my_provider_test.go`:

```go
package internal_myprovider_telephony

import (
    "context"
    "testing"

    "github.com/rapidaai/api/assistant-api/config"
    "github.com/rapidaai/pkg/commons"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewMyProviderTelephony(t *testing.T) {
    logger, _ := commons.NewApplicationLogger()
    cfg := &config.AssistantConfig{}

    provider, err := NewMyProviderTelephony(cfg, logger)
    require.NoError(t, err)
    assert.NotNil(t, provider)
}

func TestMakeCall(t *testing.T) {
    logger, _ := commons.NewApplicationLogger()
    cfg := &config.AssistantConfig{
        PublicAssistantHost: "example.com",
    }

    provider, _ := NewMyProviderTelephony(cfg, logger)

    // Test with mock credentials
    credential := &protos.VaultCredential{
        // TODO: Add mock credential
    }

    metadata, metrics, events, err := provider.MakeCall(
        auth,           // Mock auth
        "+1234567890",  // to
        "+9876543210",  // from
        1,              // assistantId
        1,              // conversationId
        credential,
        nil,
    )

    // Should fail without real credentials
    assert.Error(t, err)
    assert.Nil(t, metadata)
    assert.Nil(t, metrics)
    assert.Nil(t, events)
}
```

---

## Provider-Specific Implementation Details

### Audio Encoding Formats

Different providers use different audio encodings:

| Provider | Format       | Sample Rate   | Encoding |
| -------- | ------------ | ------------- | -------- |
| Twilio   | μ-law (PCMU) | 8000 Hz       | Base64   |
| Vonage   | Linear PCM   | 16000 Hz      | Base64   |
| Exotel   | Linear PCM   | 8000/16000 Hz | Base64   |

### Event Formats

Each provider sends events in different formats:

**Twilio (Form-encoded):**

```
POST /callback
Content-Type: application/x-www-form-urlencoded

CallStatus=answered&Timestamp=2023-01-15T10:30:00Z&...
```

**Vonage (JSON):**

```
POST /callback
Content-Type: application/json

{
  "status": "answered",
  "uuid": "abc123",
  "timestamp": "2023-01-15T10:30:00Z"
}
```

**Exotel (JSON):**

```
POST /callback
Content-Type: application/json

{
  "CallSid": "call123",
  "CallStatus": "answered",
  "CallType": "inbound"
}
```

### Connection URL Formats

Each provider has different URL structures:

**Twilio TwiML:**

```xml
<Response>
  <Connect>
    <Stream url="wss://example.com/stream" />
  </Connect>
</Response>
```

**Vonage NCCO:**

```json
[
  {
    "action": "connect",
    "endpoint": [
      {
        "type": "websocket",
        "uri": "wss://example.com/stream"
      }
    ]
  }
]
```

**Exotel HTTP:**

```json
{
  "action": "connect",
  "endpoint": {
    "type": "websocket",
    "uri": "wss://example.com/stream"
  }
}
```

---

## Best Practices

### 1. **Authentication Management**

```go
// Always extract credentials safely
if credential == nil {
    return nil, fmt.Errorf("missing credentials")
}

credMap := credential.GetValue().AsMap()
if credMap == nil {
    return nil, fmt.Errorf("invalid credential format")
}

apiKey, ok := credMap["api_key"].(string)
if !ok {
    return nil, fmt.Errorf("api_key not found or invalid type")
}
```

### 2. **Error Context**

```go
// Always wrap errors with provider context
if err != nil {
    mpt.logger.Errorf("my-provider: operation failed: %v", err)
    return fmt.Errorf("my-provider: %w", err)
}
```

### 3. **Thread Safety**

```go
// Protect shared state with mutex
mws.mu.Lock()
mws.streamId = newId
mws.mu.Unlock()

// Or use defer for guaranteed unlock
mws.mu.Lock()
defer mws.mu.Unlock()
audioData := mws.inputAudioBuffer
```

### 4. **Logging**

```go
// Log important events for debugging
mpt.logger.Debugf("my-provider: call initiated. CallID: %s", callId)
mpt.logger.Warnf("my-provider: unexpected event type: %s", eventType)
mpt.logger.Errorf("my-provider: critical error: %v", err)
```

### 5. **Metadata & Metrics**

```go
// Always return meaningful metadata
mtds := []*types.Metadata{
    types.NewMetadata("telephony.provider", "my-provider"),
    types.NewMetadata("telephony.toPhone", toPhone),
    types.NewMetadata("telephony.callId", callId),
}

// Return status metrics
metrics := []*types.Metric{
    types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Call initiated")),
}
```

### 6. **WebSocket Lifecycle**

```go
// Always check connection state
if mws.conn == nil {
    return fmt.Errorf("WebSocket not connected")
}

// Lock before sending to prevent concurrent writes
mws.mu.Lock()
defer mws.mu.Unlock()
mws.conn.WriteMessage(...)
```

---

## Troubleshooting

| Issue                          | Solution                                                         |
| ------------------------------ | ---------------------------------------------------------------- |
| Credentials not found          | Verify vault format, check credential.GetValue().AsMap()         |
| WebSocket connection fails     | Ensure WSS URL is publicly accessible, check firewall            |
| Audio not received             | Verify audio encoding/decoding matches provider format           |
| Status callbacks not triggered | Check callback URL in provider dashboard, verify public hostname |
| Call creation fails            | Check provider API limits, quota, region availability            |
| Concurrent send errors         | Use mutex to protect WebSocket writes                            |

---

## Integration Checklist

- [ ] Create provider directory and files
- [ ] Implement all Telephony interface methods
- [ ] Implement WebSocket Streamer
- [ ] Handle provider's audio encoding/decoding
- [ ] Parse provider's event format
- [ ] Extract call details from incoming requests
- [ ] Create TwiML/NCCO/custom response format
- [ ] Register provider in factory
- [ ] Add unit tests
- [ ] Test with real provider credentials
- [ ] Handle error cases gracefully
- [ ] Add logging at key points
- [ ] Document provider-specific quirks
- [ ] Update router/handlers to accept new provider constant

---

## Summary

To implement a new telephony provider:

1. **Create files**: `provider.go` and `provider_websocket.go`
2. **Implement Telephony interface**: 6 methods for call management
3. **Implement Streamer interface**: Real-time audio handling via WebSocket
4. **Register in factory**: Add case to `GetTelephony()`
5. **Test thoroughly**: Unit + integration tests with real credentials
6. **Follow patterns**: Use provider's specific formats (TwiML, NCCO, etc.)

The telephony package provides a clean abstraction for voice communications, enabling developers to add new providers while maintaining consistent behavior across all implementations.
