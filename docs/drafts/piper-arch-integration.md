# Piper on Arch Troubleshooting Notes

## Context
The assistant agent invokes Piper via the `tts` section in `assistant.toml`. On Manjaro/Arch we bundle Piper under `/opt/piper` and need to run it directly from the Go service.

## Observed Failure
- Symptom: TTS logs show `piper failed: exit status 127` when the agent handles MQTT messages.
- Piper works when run manually from the shell with `LD_LIBRARY_PATH=/opt/piper/lib`.

## Root Cause
The Go wrapper previously hard-coded Piper arguments and did not forward extra CLI arguments or environment variables defined in the TOML config. When the agent was launched via `just dev-run`, Piper inherited an empty `LD_LIBRARY_PATH`, so the binary exited with 127 (linker could not load `libonnxruntime.so`). Because we used `Run()` the underlying stderr output was swallowed, hiding the real cause.

## Fix Summary
1. Extend the `TTSConfig` struct to read `args` and `env` from `assistant.toml`.
2. Update `PiperTTS.Speak` to:
   - Merge `cfg.TTS.Args` with the default `--model` argument.
   - Pass through configured env vars in addition to the current process environment.
   - Capture combined stdout/stderr so we print the actual Piper error message.
3. Rebuild the agent so the binary picks up the updated wiring (`just build`).

## Arch/Manjaro Setup Checklist
- Install dependencies: `sudo pacman -S --needed mosquitto alsa-utils just base-devel cmake git` (or run `just deps`).
- Place the Piper files under `/opt/piper`:
  - `bin/piper`
  - `models/en_US-amy-medium.onnx`
  - `lib/` containing ONNX Runtime and other shared libs
- Update `assistant/cfg/assistant.toml`:
  ```toml
  [tts]
  bin = "/opt/piper/bin/piper"
  voice = "/opt/piper/models/en_US-amy-medium.onnx"
  args = ["--output_file", "/tmp/aurapack_tts.wav"]
  env = { LD_LIBRARY_PATH = "/opt/piper/lib" }
  ```
- Rebuild the agent: `cd assistant/ops && just build`.
- Start mosquitto for local tests: `just dev-broker` (runs in foreground).
- In another terminal, run the agent: `just dev-run`.

If Piper still fails after these steps, the agent will now print the stderr message from Piper or `aplay`. Use that output to adjust paths or install missing libraries.
