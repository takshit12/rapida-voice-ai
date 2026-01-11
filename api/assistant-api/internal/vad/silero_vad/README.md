# Silero VAD

## Initialization Logic

1. Resolve the Silero ONNX model path:

   - If the environment variable `SILERO_MODEL_PATH` is set, use its value.
   - Otherwise, use the bundled model path:
     ```
     models/silero_vad_20251001.onnx
     ```

2. Load the Silero ONNX model from the resolved path.

   - If loading fails, return an error and abort initialization.

3. Resolve the speech detection threshold:

   - Read `microphone.vad.threshold` from configuration.
   - If not provided, default to `0.5`.

4. Initialize the Silero detector using:

   - The loaded ONNX model
   - The resolved detection threshold

5. Create an internal audio processing configuration:

   - Sample rate: 16 kHz
   - Channels: mono
   - Audio format: linear PCM converted to `float32`

6. Store the activity callback function used to report detected speech.

7. Return the initialized `SileroVAD` instance.

---

## Audio Processing Logic (`Process`)

1. Receive an incoming chunk of audio data.

2. Validate and normalize the audio format:

   - If the input audio is not 16 kHz mono, resample it accordingly.
   - If resampling fails:
     - Log the error
     - Return the error

3. Convert the normalized audio samples to `float32`.

   - If conversion fails:
     - Log the error
     - Return the error

4. Execute the Silero VAD detector on the converted audio buffer.

5. Collect all speech segments detected by the model.

6. If **no speech segments** are detected:

   - Do not invoke the activity callback
   - Return `nil`

7. If **one or more speech segments** are detected:

   - Identify the earliest segment start time.
   - Identify the latest segment end time.
   - Merge all detected segments into a single speech activity window.

8. Create an `internal_vad.VadResult` using the merged start and end times.

9. Invoke the activity callback with the generated VAD result.

10. Return `nil` unless an error occurred during processing.

---

## Identifier Logic (`Name`)

1. Return the constant VAD identifier:
   ```
   silero_vad
   ```

---

## Cleanup Logic (`Close`)

1. Destroy the Silero detector instance.
2. Release all associated resources.
3. Return `nil`.

---

## Runtime Constraints and Behavior

- Audio must be processed as a continuous stream.
- The detector is stateful and maintains context across `Process` calls.
- A single `SileroVAD` instance must be reused for real-time audio streams.
- Detection accuracy depends on:
  - Microphone quality
  - Environmental noise
  - Correct audio resampling and conversion
