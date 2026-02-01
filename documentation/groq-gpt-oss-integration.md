# Groq GPT-OSS Integration Notes

## Architecture

Groq models are accessed through the **OpenAI-compatible API** (`https://api.groq.com/openai/v1`). The integration reuses the OpenAI caller (`api/integration-api/internal/caller/openai/`) with a custom base URL from the credential configuration.

## GPT-OSS Raw HTTP Path

GPT-OSS models (`openai/gpt-oss-120b`, `openai/gpt-oss-20b`) use a **raw HTTP/1.1 implementation** instead of the OpenAI Go SDK. This is implemented in `rawStreamGPTOSS()` in `llm.go`.

### Why Not the SDK?

The OpenAI Go SDK v1.12.0 uses HTTP/2 by default (Go's `net/http` upgrades TLS connections to HTTP/2). Groq's GPT-OSS model handler has an HTTP/2 incompatibility that causes it to return empty responses with `prompt_tokens: 0`. The same request over HTTP/1.1 works correctly.

This was discovered after exhaustive debugging that ruled out:
- JSON serialization differences (SDK `apijson` vs `json.Marshal`)
- Extra HTTP headers (X-Stainless-*, OpenAI-*)
- SDK-injected null fields from union types
- Parameter values (temperature, top_p, reasoning_effort, etc.)
- Message format (content parts array vs simple string)
- Message roles (system vs user)

### How It Works

```
StreamChatCompletion()
├── model starts with "openai/gpt-oss" ?
│   └── YES → rawStreamGPTOSS()     ← raw net/http, HTTP/1.1
│       ├── json.Marshal request body
│       ├── POST with only Authorization + Content-Type headers
│       ├── TLSNextProto: {} (disables HTTP/2)
│       └── bufio.Scanner SSE parsing
└── NO → OpenAI SDK NewStreaming()   ← standard path for all other models
```

### GPT-OSS Streaming Format

GPT-OSS has a non-standard SSE format compared to regular OpenAI models:

```
# 1. Reasoning tokens (not in delta.content!)
data: {"choices":[{"delta":{"reasoning":"thinking text","channel":"analysis"}}]}

# 2. Content tokens
data: {"choices":[{"delta":{"content":"actual response"}}]}

# 3. Final chunk with usage
data: {"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{...}}

data: [DONE]
```

The `delta.reasoning` field contains the model's chain-of-thought. These tokens are NOT sent to the user — only `delta.content` is streamed to the UI.

### Fixed Parameters

For GPT-OSS, these parameters are hardcoded (matching Groq's working examples):

| Parameter | Value | Notes |
|-----------|-------|-------|
| `temperature` | 1 | Lower values may cause empty responses |
| `top_p` | 1 | |
| `max_completion_tokens` | 8192 | |
| `reasoning_effort` | "medium" | Supports "low", "medium", "high" |

## Non-GPT-OSS Models (Llama, Mixtral, etc.)

All other Groq models use the standard OpenAI SDK path. These models work correctly with HTTP/2 and the SDK's serialization.

Additional Groq-specific handling in the SDK path:
- **Header stripping**: `X-Stainless-*` and `OpenAI-*` headers are removed for non-OpenAI hosts in the HTTP middleware (`openai.go`)
- **Clean params struct**: GPT-OSS models get a fresh `ChatCompletionNewParams` with only required fields (avoids SDK serializing null union fields)
- **User message format**: Text-only user messages use simple string format instead of content parts array

## Credential Configuration

Groq credentials are stored with:
- `key`: Groq API key
- `url`: `https://api.groq.com/openai/v1` (custom base URL)

## Debugging

The HTTP middleware in `openai.go` logs:
- All request headers (excluding Authorization)
- Full request body JSON
- Response status code
- Error response bodies (for 4xx/5xx)

For GPT-OSS raw path, `rawStreamGPTOSS:` prefixed logs show:
- Request body
- Response status and HTTP protocol version
- Each SSE chunk (including reasoning tokens)
- Stream completion summary with token counts

## Future Considerations

- If Groq fixes their HTTP/2 handling for GPT-OSS, the raw HTTP path could be removed and GPT-OSS could use the SDK like other models
- The `delta.reasoning` tokens could be exposed in the UI for transparency (currently only logged)
- New GPT-OSS models should be added to the `openai/gpt-oss` prefix check in `llm.go`
