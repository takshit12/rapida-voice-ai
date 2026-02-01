# Release Notes - February 2, 2026

## Overview

This release fixes **GPT-OSS models (120B/20B) returning empty responses** when accessed via Groq through the OpenAI Go SDK, and adds **Groq as a text/LLM provider** with full model support including GPT-OSS, Llama 4, Kimi K2, Qwen 3, and others.

---

## Bug Fix: GPT-OSS Empty Responses on Groq

### Problem

GPT-OSS-120B (and GPT-OSS-20B) on Groq returned empty responses with `prompt_tokens: 0` when called through the OpenAI Go SDK v1.12.0. The model responded in ~130ms (too fast for actual processing) with 2 empty SSE chunks and no content. Other models (Llama 3.3 70B, Mixtral) worked fine through the same code path.

The same request worked perfectly via curl from Groq's playground.

### Root Cause

**The OpenAI Go SDK's HTTP/2 transport causes Groq's GPT-OSS model to return empty responses.**

Go's `net/http` client defaults to HTTP/2 over TLS, while curl defaults to HTTP/1.1. Groq's GPT-OSS model handler appears to have an incompatibility with HTTP/2 connections that causes it to return immediately without processing the input (evidenced by `prompt_tokens: 0`).

This was confirmed by systematically eliminating all other variables:

| Attempt | Change | Result |
|---------|--------|--------|
| 1 | Fix user message format (content parts array → simple string) | Still empty |
| 2 | Fix tool_choice sent without tools | Still empty |
| 3 | Strip frequency_penalty/presence_penalty when zero | Still empty |
| 4 | Convert system messages to user messages (per Groq docs) | Still empty |
| 5 | Strip temperature/top_p for GPT-OSS | Still empty |
| 6 | Strip ALL optional params (only model + messages + stream) | Still empty |
| 7 | Match exact working Groq params (temperature=1, top_p=1, etc.) | Still empty |
| 8 | Use fresh params struct (eliminate SDK serialized nulls) | Still empty |
| 9 | Strip SDK headers (X-Stainless-*, OpenAI-*) for Groq | Still empty |
| **10** | **Bypass SDK entirely: raw HTTP/1.1 with net/http** | **Working** |

### Fix

For GPT-OSS models, bypass the OpenAI Go SDK entirely and use a raw `net/http` implementation that:

1. **Forces HTTP/1.1** — disables Go's HTTP/2 by setting `TLSNextProto` to empty map
2. **Sends only `Authorization` and `Content-Type` headers** — no SDK-injected headers
3. **Uses standard `json.Marshal`** — no SDK `apijson` serialization
4. **Parses SSE with `bufio.Scanner`** — no SDK streaming decoder

All other models continue using the OpenAI SDK as before.

### Technical Details

**GPT-OSS streaming response format** (different from standard OpenAI):

```
# Reasoning tokens come first in delta.reasoning with channel:"analysis"
{"choices":[{"delta":{"reasoning":"thinking...","channel":"analysis"}}]}

# Then content tokens in delta.content
{"choices":[{"delta":{"content":"नमस्ते!"}}]}

# Final chunk with usage
{"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":2294,"completion_tokens":89,...,"completion_tokens_details":{"reasoning_tokens":65}}}
```

**Key parameters for GPT-OSS on Groq:**
- `temperature: 1` (required, lower values may cause issues)
- `top_p: 1`
- `max_completion_tokens: 8192`
- `reasoning_effort: "medium"` (supports "low", "medium", "high")
- System role messages work (despite Groq docs suggesting otherwise)

### Files Modified

| File | Change |
|------|--------|
| `api/integration-api/internal/caller/openai/llm.go` | Added `rawStreamGPTOSS()` function; GPT-OSS models route to raw HTTP path in `StreamChatCompletion()` |
| `api/integration-api/internal/caller/openai/openai.go` | Strip SDK headers for non-OpenAI hosts; log HTTP request/response details |

### Commits

| Commit | Description |
|--------|-------------|
| `2a0a6b3` | Fix user message format to simple string for Groq compatibility |
| `be6c0e7` | Don't send empty tool_choice string |
| `f864d8d` | Don't send tool_choice when no tools defined |
| `2d26efe` | Strip frequency_penalty/presence_penalty when zero |
| `c6c46be` | Add HTTP request body logging middleware |
| `bdeef2e` | Convert system→user messages for reasoning models (later reverted) |
| `af0753f` | Add HTTP response logging, strip temperature/top_p |
| `7c451d0` | Strip ALL optional params for GPT-OSS |
| `c967c1e` | Match exact working Groq params, add header logging |
| `20b90fe` | Fresh params struct, strip SDK headers for Groq |
| `8d7aff5` | **Final fix:** Bypass SDK with raw HTTP/1.1 streaming |

---

## New Feature: Groq Text/LLM Provider

### Added Models

| Model ID | Name | Description |
|----------|------|-------------|
| `groq/openai/gpt-oss-120b` | GPT OSS 120B | OpenAI flagship open-weight MoE model, 131K context |
| `groq/openai/gpt-oss-20b` | GPT OSS 20B | Fast OpenAI open-weight model, ~1000 T/s |
| `groq/llama-3.3-70b-versatile` | Llama 3.3 70B | Most capable Llama model for complex tasks |
| `groq/llama-3.1-8b-instant` | Llama 3.1 8B | Fastest model for real-time applications |
| `groq/meta-llama/llama-4-scout-17b-16e-instruct` | Llama 4 Scout 17B | Meta MoE model (109B total, 17B active), multimodal |
| `groq/meta-llama/llama-4-maverick-17b-128e-instruct` | Llama 4 Maverick 17B | Meta large MoE (128 experts), multimodal |
| `groq/moonshotai/kimi-k2-instruct` | Kimi K2 | Moonshot AI 1T MoE, excels at tool use and coding |
| `groq/qwen/qwen-3-32b` | Qwen 3 32B | Alibaba Qwen 3, 131K context, ~662 T/s |
| `groq/llama-3.2-90b-vision-preview` | Llama 3.2 90B Vision | Vision-capable Llama for multimodal tasks |
| `groq/llama-3.2-11b-vision-preview` | Llama 3.2 11B Vision | Compact vision-capable Llama |
| `groq/mixtral-8x7b-32768` | Mixtral 8x7B (32K) | Mixture of experts with 32K context |
| `groq/gemma2-9b-it` | Gemma 2 9B | Google Gemma 2 instruction-tuned |

### Configuration Options

- **Temperature** (0-2): Sampling temperature
- **Top P** (0-1): Nucleus sampling
- **Frequency Penalty** (-2 to 2): Penalize repeated tokens
- **Presence Penalty** (-2 to 2): Penalize tokens already in text
- **Max Completion Tokens**: Upper bound for generated tokens
- **Tool Choice**: none / auto / required
- **Response Format**: text / json_object
- **Reasoning Effort** (GPT-OSS only): low / medium / high
- **Stop Sequences**: Up to 4 comma-separated stop strings

### Files Added/Modified

| File | Change |
|------|--------|
| `ui/src/app/components/providers/text/groq/constants.ts` | Model list, default options, validation |
| `ui/src/app/components/providers/text/groq/index.tsx` | Configuration UI component |
| `ui/src/app/components/providers/text/provider.tsx` | Register Groq in provider switch |
| `ui/src/providers/provider.development.json` | Add Groq provider entry |
| `ui/src/providers/provider.production.json` | Add Groq provider entry |

---

## Lessons Learned

1. **HTTP/2 incompatibility with some API providers**: Go's default HTTP/2 can cause silent failures with providers that have HTTP/2 bugs. When a model returns `prompt_tokens: 0`, the issue may be at the transport level, not the request format.

2. **OpenAI SDK serialization quirks**: The SDK's `apijson` serializer with `paramUnion` embedded types may serialize union fields as `null` even when not set, creating unexpected JSON fields. Using a fresh struct avoids this.

3. **Systematic debugging approach**: When curl works but SDK doesn't with identical JSON, the problem is in headers, HTTP version, or transport-level behavior — not the request body.

4. **Groq GPT-OSS specifics**: The model sends reasoning tokens in a separate `delta.reasoning` field (not `delta.content`) with `channel: "analysis"`. Standard OpenAI SDK streaming decoders don't parse this field.

---

## Verification

After deployment, logs confirmed:
- `rawStreamGPTOSS: response status=200 proto=HTTP/1.1` (HTTP/1.1 confirmed)
- `rawStreamGPTOSS: stream complete. chunks=81 content_len=101 prompt_tokens=2294 completion_tokens=89`
- Model responds in Hindi/Hinglish following the system prompt correctly
- Reasoning tokens visible in logs (65 reasoning tokens before 24 content tokens)

---

## Contributors

- Debugging and implementation assisted by Claude Opus 4.5
