# Release Notes - January 31, 2026

## Overview

This release fixes **critical voice pipeline failures** on Railway that prevented phone calls from producing transcripts and audio responses. Three separate issues were identified and resolved across four commits.

---

## Bug Fixes

### 1. Silero VAD Model Missing in Docker Container

**Problem:** Voice Activity Detection (VAD) failed to initialize on every phone call with:
```
error wile intializing vad failed to create silero detector:
  failed to create session: Load model from
  /app/api/assistant-api/internal/vad/internal/silero_vad/models/silero_vad_20251001.onnx failed.
  File doesn't exist
```

**Root Cause:** Two compounding issues:
- `.gitignore` excludes `*.onnx` files, so the Silero VAD model was never committed to the repository or included in Docker builds
- The model path resolver (`silero_vad.go:217`) uses `runtime.Caller(0)` which returns the build-time source path `/app/...`, but the compiled binary runs from `/opt/apps` in the runtime container

**Fix:**
- Download the Silero VAD v5 ONNX model from the official repository during Docker build
- Copy it to `/opt/apps/models/silero_vad.onnx` in the runtime stage
- Set `SILERO_MODEL_PATH` environment variable to bypass the `runtime.Caller(0)` path resolution

**Impact:** VAD now initializes successfully. Voice activity detection, interruption handling, and the full audio pipeline are operational.

### 2. File Storage Permission Denied

**Problem:** Call recordings failed to save with:
```
unable to create complete path, err mkdir /app: permission denied
```

**Root Cause:** `ASSET_STORE__STORAGE_PATH_PREFIX` was set to `/app/rapida-data/assets/workflow`. The runtime container uses `/opt/apps` as its working directory and runs as non-root user `rapida-app`, which cannot create `/app`.

**Fix:** Changed storage path prefix to `/opt/apps/rapida-data/assets/workflow` and ensured the directory is created during Docker build with correct ownership.

**Impact:** Call recordings are now saved successfully.

### 3. Phone Deployment Audio Config Errors Silently Ignored

**Problem:** When creating a phone deployment, errors from `createAssistantDeploymentAudio()` were silently ignored. If audio config creation failed, the deployment was returned without STT/TTS configuration, resulting in calls with no audio processing.

**Fix:** Added proper error checking and propagation for audio config creation. Also fixed the `IsPhoneDeploymentEnable()` check which previously returned nil instead of an error.

### 4. Twilio WebSocket URL Pointing to Invalid Domain

**Problem:** `PUBLIC_ASSISTANT_HOST` was set to `ngork.local.dev` (a development placeholder). Twilio uses this to build WebSocket media stream URLs (`wss://`) and status callback URLs (`https://`). Since the domain was unreachable, calls connected but no audio was ever streamed.

**Fix:** Set `PUBLIC_ASSISTANT_HOST` to the actual Railway domain `assistant-api-production-2ea9.up.railway.app`.

---

## Files Modified

| Commit | Files | Description |
|--------|-------|-------------|
| `8d055ba` | `docker/assistant-api/Dockerfile`, `docker/assistant-api/.assistant.env` | Add VAD model download, fix storage path |
| `4759a83` | `api/assistant-api/internal/services/assistant/assistant.deployment.impl.service.go`, `api/assistant-api/api/talk/inbound-call.go` | Handle audio config errors, fix phone deployment check |
| `f63bba1` | `docker/assistant-api/.assistant.env` | Fix PUBLIC_ASSISTANT_HOST domain |

---

## Dockerfile Changes

### Builder Stage
```dockerfile
# Download Silero VAD model (v5, required by silero-vad-go v0.2.1)
RUN mkdir -p /opt/models && \
    wget -q -O /opt/models/silero_vad.onnx \
    https://github.com/snakers4/silero-vad/raw/master/src/silero_vad/data/silero_vad.onnx
```

### Runtime Stage
```dockerfile
# Copy VAD model
COPY --from=builder /opt/models/silero_vad.onnx /opt/apps/models/silero_vad.onnx

# Create required directories (updated)
RUN mkdir -p /opt/apps/assets /opt/apps/rapida-data/assets/workflow /opt/apps/models /var/log/go-app && \
    chown -R rapida-app:rapida-app /opt/apps /var/log/go-app

ENV SILERO_MODEL_PATH="/opt/apps/models/silero_vad.onnx"
```

---

## Known Remaining Issue

### OpenSearch Unreachable
```
error while bulk operation to opensearch got error dial tcp: lookup opensearch on [fd12::10]:53: no such host
```
No OpenSearch service is deployed on Railway. This only affects conversation search indexing and does not block voice calls. Can be addressed by deploying an OpenSearch service if full-text search is needed.

---

## Verification

After deployment, Railway logs confirmed:
- `listen.initializeVAD took 226.95ms` (previously failed)
- `deepgram-stt: connection established` (STT working)
- No more "mkdir /app: permission denied" errors
- No more "illegal input audio transformer" errors

---

## Contributors

- Implementation and debugging assisted by Claude Opus 4.5
