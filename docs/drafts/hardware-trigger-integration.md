# Draft: Voice Trigger Integration (VR3 → Go → Whisper → LLM)

## Status
Draft, 2025-11-03

## Context
Aurapack's **assistant** service must stay idle until a user intentionally wakes it.
Running Whisper continuously on a Raspberry Pi wastes CPU and power. We need a **hardware trigger** that detects wake words locally and signals the assistant to start the full voice-processing pipeline.

The **Elechouse Voice Recognition Module V3 (VR3)** is used as that offline, low-power trigger.
It recognizes a small set of trained phrases and notifies the **Go assistant service** via serial (UART). From there, the assistant starts audio capture, transcription, and reasoning.

This document describes the software integration, focusing on communication and orchestration rather than wiring or electrical details.

## Decision

### Trigger → Assistant integration

The VR3 continuously listens for predefined keywords and, when one is detected, sends a data packet to the Raspberry Pi through serial (e.g., `/dev/ttyAMA0`).
The assistant, implemented in Go, keeps a lightweight event loop open on that port using the **go.bug.st/serial** library.

The VR3 uses a specific protocol structure:
```
| Header (0xAA) | Length | Command | Data | End (0x0A) |
```

When voice is recognized, it sends:
```
| 0xAA | 0x0C | 0x0D | <record_id> | 0x0A |
```

Where `0x0D` is the "voice recognized" command and `record_id` identifies which trained phrase was detected.

Simplified structure:

```go
// trigger/vr3.go
package trigger

import (
    "context"
    "fmt"
    "io"
    "go.bug.st/serial"
)

type VR3Config struct {
    Port     string
    BaudRate int
}

func Listen(ctx context.Context, cfg VR3Config, onTrigger func(recordID byte)) error {
    mode := &serial.Mode{
        BaudRate: cfg.BaudRate,
        Parity:   serial.NoParity,
        DataBits: 8,
        StopBits: serial.OneStopBit,
    }

    port, err := serial.Open(cfg.Port, mode)
    if err != nil {
        return fmt.Errorf("failed to open port: %w", err)
    }
    defer port.Close()

    buf := make([]byte, 128)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        n, err := port.Read(buf)
        if err != nil {
            if err == io.EOF {
                return nil
            }
            return fmt.Errorf("read error: %w", err)
        }

        // Parse VR3 protocol frames
        for i := 0; i < n; i++ {
            if buf[i] == 0xAA && i+4 < n {
                if buf[i+1] == 0x0C && buf[i+2] == 0x0D && buf[i+4] == 0x0A {
                    recordID := buf[i+3]
                    fmt.Printf("[VR3] Detected command: %d\n", recordID)
                    onTrigger(recordID)
                    i += 4
                }
            }
        }
    }
}
```

In the main assistant process:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

cfg := trigger.VR3Config{
    Port:     "/dev/ttyAMA0",
    BaudRate: 115200,
}

go trigger.Listen(ctx, cfg, func(id byte) {
    if id == 0x01 {
        fmt.Println("[Assistant] Wake word received, starting capture")
        StartVoiceSession()
    }
})
```

This keeps the assistant passive until the VR3 sends a signal, consuming almost no resources in standby.

### Voice session flow

Once awakened, the assistant performs a full processing cycle:

```
VR3 trigger → Go service → Whisper (STT) → LLM normalizer → Intent JSON → MQTT → Orchestrator → Device adapters → Piper (TTS)
```

- **Trigger layer**: purely event based, driven by hardware.
- **STT layer**: runs Whisper as a subprocess, records until silence detected.
- **Normalizer (LLM)**: converts free text to structured JSON describing the action.
- **Orchestrator**: validates and routes intents to device adapters.
- **Actuation**: publishes the resulting intent through MQTT.
- **Feedback**: generates a spoken confirmation using Piper.

The Go service manages all subprocess orchestration, using channels and context cancellation to keep lifetimes under control.

### Configuration and runtime

All behavior is configured in `assistant.toml`:

```toml
[trigger]
enabled = true
port = "/dev/ttyAMA0"
baud_rate = 115200

[stt]
bin = "/opt/aurapack/assistant/bin/whisper"
model = "/opt/aurapack/assistant/models/ggml-base.en.bin"

[tts]
bin = "/opt/piper/bin/piper-wrapper"
voice = "/opt/piper/models/en_US-amy-medium.onnx"
```

The assistant runs continuously under systemd. When idle, only the serial listener remains active. When triggered, it launches Whisper, collects text, calls the LLM (local network or embedded), and dispatches results to MQTT.

### Future directions

- Add configurable debounce and cooldown between triggers.
- Integrate additional trigger types (e.g., GPIO button, hotword microcontroller).
- Expose simple `/system/triggers` endpoint for monitoring trigger activity.
- Implement MQTT broadcast of trigger events for distributed agents.

## Expected Result

- The assistant remains effectively silent until a local, offline trigger occurs.
- CPU use near zero in standby.
- Entire flow from hardware signal to action stays within the Go service.
- No external scripts or redundant background processes.
- The approach is architecture neutral and reproducible on both Arch and Raspberry.

## Implementation Notes

This document represents the current design direction but has not been fully validated in production. Several aspects require verification:

- The protocol parsing logic assumes the VR3 default frame structure based on available documentation. This should be tested with the actual hardware.
- The serial library choice (go.bug.st/serial) has been selected based on maintenance status and cross-platform support, but alternatives may be considered during implementation.
- The interaction between multiple concurrent triggers has not been fully designed. Debounce logic will be added based on testing.

This document will be promoted to a formal ADR once the implementation has been tested and validated on actual hardware.

