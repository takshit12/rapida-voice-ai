# Release Notes - January 29, 2026

## Overview

This release adds **Groq LLM provider support** to rapida-voice-ai and fixes a critical **WebSocket connection issue** that was preventing the assistant preview from working in production.

---

## New Features

### Groq LLM Provider Integration

Groq is now available as a text/LLM provider, offering blazing fast AI inference with LPU (Language Processing Unit) technology. Groq uses an **OpenAI-compatible API**, which allowed us to leverage existing infrastructure.

#### Supported Models

| Model ID | Human Name | Description |
|----------|------------|-------------|
| `groq/llama-3.3-70b-versatile` | Llama 3.3 70B Versatile | Most capable Llama model for complex tasks |
| `groq/llama-3.1-8b-instant` | Llama 3.1 8B Instant | Fastest model for real-time applications |
| `groq/llama-3.2-90b-vision-preview` | Llama 3.2 90B Vision | Vision-capable Llama model for multimodal tasks |
| `groq/llama-3.2-11b-vision-preview` | Llama 3.2 11B Vision | Compact vision-capable Llama model |
| `groq/mixtral-8x7b-32768` | Mixtral 8x7B (32K) | Mixture of experts with 32K context window |
| `groq/gemma2-9b-it` | Gemma 2 9B | Google Gemma 2 instruction-tuned model |
| `groq/llama-guard-3-8b` | Llama Guard 3 8B | Safety and moderation model |

#### Configurable Parameters

- **Temperature** (0-2): Controls randomness in output
- **Top P** (0-1): Nucleus sampling parameter
- **Frequency Penalty** (-2 to 2): Penalizes token repetition based on frequency
- **Presence Penalty** (-2 to 2): Penalizes tokens based on presence in text
- **Max Completion Tokens**: Upper bound for generated tokens
- **Stop Sequences**: Up to 4 sequences where generation stops
- **Tool Choice**: none, auto, or required
- **Response Format**: JSON object for structured output

---

## Bug Fixes

### WebSocket Connection Error in Production

**Problem:** The assistant preview was failing with the error:
```
Websocket transport constructed with non-https:// or http:// host
```

**Root Cause:** The `@improbable-eng/grpc-web` library (used by `@rapidaai/react`) requires **full URLs with http:// or https:// scheme** for WebSocket transport. The library converts:
- `https://` → `wss://`
- `http://` → `ws://`

The production config was using **relative paths** (`/api/assistant`) instead of full URLs, which cannot be converted to WebSocket URLs.

**Solution:** Updated `ui/src/configs/config.production.json` to use full URLs for all connection endpoints.

---

## Technical Implementation Details

### Files Created

#### Backend - Groq Caller Package

**`api/integration-api/internal/caller/groq/groq.go`**
```go
package internal_groq_callers

import (
    "errors"
    "github.com/openai/openai-go"
    "github.com/openai/openai-go/option"
    internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
    "github.com/rapidaai/pkg/commons"
)

type Groq struct {
    logger     commons.Logger
    credential internal_callers.CredentialResolver
}

var (
    GROQ_BASE_URL = "https://api.groq.com/openai/v1"
    API_KEY       = "key"
)

func NewGroq(logger commons.Logger, credential internal_callers.CredentialResolver) *Groq {
    return &Groq{
        logger:     logger,
        credential: credential,
    }
}

func (g *Groq) GetClient() (*openai.Client, error) {
    credentials := g.credential()
    apiKey, ok := credentials[API_KEY]
    if !ok {
        return nil, errors.New("unable to resolve the credential")
    }
    clt := openai.NewClient(
        option.WithAPIKey(apiKey.(string)),
        option.WithBaseURL(GROQ_BASE_URL),
    )
    return &clt, nil
}
```

**`api/integration-api/internal/caller/groq/llm.go`**
- Implements `LargeLanguageCaller` interface
- `GetChatCompletion()` - Synchronous chat completion
- `StreamChatCompletion()` - Streaming chat completion with SSE
- Handles Groq-specific parameters (temperature, top_p, frequency_penalty, etc.)
- Uses OpenAI SDK since Groq API is OpenAI-compatible

**`api/integration-api/internal/caller/groq/verify-credential.go`**
- Credential verification using `llama-3.3-70b-versatile` model
- Tests API key validity with a simple chat completion request

#### Frontend - Groq Text Provider

**`ui/src/app/components/providers/text/groq/constants.ts`**
- `GROQ_TEXT_MODEL` array with all supported models
- `GetGroqTextProviderDefaultOptions()` - Sets default parameter values
- `ValidateGroqTextProviderDefaultOptions()` - Validates user input

**`ui/src/app/components/providers/text/groq/index.tsx`**
- `ConfigureGroqTextProviderModel` component
- Model dropdown selector
- Parameter configuration popover with sliders and inputs
- Real-time parameter validation

### Files Modified

#### Backend

**`pkg/clients/integration/integration_client.go`**

Added Groq routing to use OpenAI client (since API is compatible):

```go
// In Chat method
case "groq":
    // Groq uses OpenAI-compatible API, reuse OpenAI client
    return client.openAiClient.Chat(client.WithAuth(c, auth), request)

// In StreamChat method
case "groq":
    return client.openAiClient.StreamChat(client.WithAuth(c, auth), request)

// In VerifyCredential method
case "groq":
    groqCaller := internal_groq_callers.NewLargeLanguageCredentialVerifier(
        client.logger,
        internal_callers.ResolveToMap(request.GetCredential()),
    )
    return groqCaller.Verify(c)
```

**`api/integration-api/internal/caller/openai/openai.go`**

Added custom base URL support for credential-based routing:

```go
func (openAI *OpenAI) GetClient() (*openai.Client, error) {
    credentials := openAI.credential()
    cx, ok := credentials[API_KEY]
    if !ok {
        return nil, errors.New("unable to resolve the credential")
    }

    opts := []option.RequestOption{
        option.WithAPIKey(cx.(string)),
    }

    // Support custom base URL from credentials
    if baseURL, ok := credentials[API_URL]; ok && baseURL != "" {
        openAI.logger.Debugf("Using custom base URL: %s", baseURL)
        opts = append(opts, option.WithBaseURL(baseURL.(string)))
    }

    clt := openai.NewClient(opts...)
    return &clt, nil
}
```

#### Frontend

**`ui/src/app/components/providers/text/index.tsx`**

Added Groq imports and switch cases:

```typescript
import { ConfigureGroqTextProviderModel } from '@/app/components/providers/text/groq';
import {
  GetGroqTextProviderDefaultOptions,
  ValidateGroqTextProviderDefaultOptions,
} from '@/app/components/providers/text/groq/constants';

// In GetDefaultTextProviderConfigIfInvalid
case 'groq':
  return GetGroqTextProviderDefaultOptions(parameters);

// In ValidateTextProviderDefaultOptions
case 'groq':
  return ValidateGroqTextProviderDefaultOptions(parameters);

// In TextProviderConfigComponent
case 'groq':
  return (
    <ConfigureGroqTextProviderModel
      parameters={parameters}
      onParameterChange={onChangeParameter}
    />
  );
```

**`ui/src/providers/provider.production.json`** and **`provider.development.json`**

Added Groq provider configuration:

```json
{
    "code": "groq",
    "name": "Groq",
    "description": "Blazing fast AI inference with LPU technology. Supports Llama, Mixtral, and Gemma models with ultra-low latency.",
    "image": "https://cdn.brandfetch.io/id2S-_5WdS/theme/dark/logo.svg",
    "featureList": [
        "text",
        "external"
    ],
    "configurations": [
        {
            "name": "key",
            "type": "string",
            "label": "API Key"
        },
        {
            "name": "url",
            "type": "string",
            "label": "Base URL",
            "default": "https://api.groq.com/openai/v1"
        }
    ],
    "humanname": "Groq",
    "website": "https://groq.com"
}
```

**`ui/src/configs/config.production.json`**

Fixed WebSocket connection issue by using full URLs:

```json
{
  "connection": {
    "assistant": "https://rapida-voice-ai-production.up.railway.app/api/assistant",
    "web": "https://rapida-voice-ai-production.up.railway.app/api/web",
    "endpoint": "https://rapida-voice-ai-production.up.railway.app/api/endpoint",
    "media": "https://rapida-voice-ai-production.up.railway.app/api/assistant"
  }
}
```

---

## Deployment

### Railway Services Affected

| Service | Changes |
|---------|---------|
| `ui` | Config changes, new Groq components |
| `integration-api` | New Groq caller package |
| `web-api` | Updated integration client routing |

### Deployment Steps

1. All changes pushed to `main` branch
2. Railway automatically detected changes and triggered builds
3. Services deployed in order: integration-api → web-api → ui

### Commits

| Commit | Description |
|--------|-------------|
| `46cb226` | Add Groq LLM provider support |
| `b6096d0` | Fix media server URL for WebSocket connections (partial fix) |
| `0adcffe` | Fix media URL to use nginx proxy for WebSocket connections |
| `0ec875f` | Fix WebSocket connection by using full URLs in production config |

---

## Testing

### Groq Integration Test

```bash
curl https://api.groq.com/openai/v1/chat/completions \
  -H "Authorization: Bearer $GROQ_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.3-70b-versatile",
    "messages": [{"role": "user", "content": "Hi"}]
  }'
```

### UI Testing Steps

1. Navigate to Credentials → Add new credential for Groq
2. Create/Edit Assistant → Select Groq as text provider
3. Choose a model (e.g., Llama 3.3 70B Versatile)
4. Configure parameters (temperature, etc.)
5. Save and test in Preview

---

## Known Issues Resolved

### Issue: "Websocket transport constructed with non-https:// or http:// host"

**Symptoms:**
- Assistant preview fails to connect
- Console shows WebSocket error
- Both text and voice modes affected

**Investigation:**
1. Initially thought `media` config was empty - fixed but error persisted
2. Tried routing through nginx proxy - still failed
3. Deep dive into `@improbable-eng/grpc-web` library revealed the real issue

**Root Cause Discovery:**
- Development config uses full URLs: `http://localhost:8080`
- Production config used relative paths: `/api/assistant`
- The grpc-web library validates URLs and requires http:// or https:// prefix
- See: https://github.com/improbable-eng/grpc-web/issues/622

**Final Fix:**
Changed all production connection endpoints from relative paths to full URLs.

---

## Architecture Notes

### Why Groq Works with OpenAI Client

Groq provides an OpenAI-compatible API at `https://api.groq.com/openai/v1`. This means:

1. Same request/response format as OpenAI
2. Same SDK can be used with different base URL
3. Minimal code changes required - just route "groq" to OpenAI client with Groq's base URL

### WebSocket Transport in @rapidaai/react

The `@rapidaai/react` package uses:
- `@improbable-eng/grpc-web` for gRPC-Web communication
- WebSocket transport for bidirectional streaming (voice, real-time updates)

The WebSocket transport requires full URLs because:
1. It needs to convert `https://` → `wss://` for secure WebSocket
2. Relative paths cannot be converted to WebSocket protocol
3. This is per WebSocket RFC specification

---

## Future Considerations

1. **Environment Variables**: Consider using environment variables for production URLs to avoid hardcoding
2. **Groq Vision Models**: The vision-capable models (llama-3.2-90b-vision, llama-3.2-11b-vision) could be integrated with multimodal features
3. **Rate Limiting**: Groq has rate limits - consider adding retry logic with exponential backoff
4. **Model Updates**: Groq regularly adds new models - keep the model list updated

---

## Contributors

- Implementation and debugging assisted by Claude Opus 4.5
