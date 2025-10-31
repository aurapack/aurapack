# ADR-0001 - Asistant Deployment and Onboarding

## Status
Proposed — 2025-10-30

## Context
The **assistant** node is the local voice and control interface in the Aurapack architecture.  
It captures audio, performs local speech-to-text (STT), text-to-speech (TTS), and communicates with other nodes via MQTT.  
During development it runs on Linux; in production, it targets low-power devices such as a Raspberry Pi or similar SBC.  
We need a reproducible, ergonomic, and self-contained way to install and operate it.

## Decision

### 1. Runtime structure
```
aurapack/
  assistant/
    agent/
      main.go
    cfg/
      assistant.toml
    ops/
      install.sh
      assistant.service
      justfile
```

- `assistant.toml`: main configuration (MQTT broker, STT/TTS paths, language, voice, etc.)
- `install.sh`: installs binaries, creates system user `aurapack`, sets up directories under `/opt/aurapack/assistant/`, and enables the `assistant.service`.
- `assistant.service`: managed by systemd with automatic restart.
- `justfile`: local build and run commands.

### Reproducible deployment
- The build process (`just package`) generates a versioned tarball:
  ```
  aurapack-assistant_<version>.tar.gz
  ```
  containing binaries, config, and systemd units.
- Deployment on target device:
  ```
  curl -sL https://auralpack.io/setup-assistant.sh | sudo bash
  ```
  The script downloads the tarball, installs it to `/opt/aurapack/assistant/`, and enables the service.

### Configuration and onboarding
- On first boot, if Wi-Fi or broker configuration is missing:
  - Launches in **Hotspot Mode**:
    - Starts `hostapd` (SSID `Aurapack-Setup`)
    - Assigns IP `192.168.50.1` via `dnsmasq`
    - Serves a lightweight setup dashboard on `http://192.168.50.1`
  - User connects from a phone/laptop and configures:
    - Wi-Fi credentials
    - Language and voice
    - Broker address (`aurapack.local` default)
  - On save:
    - Writes `assistant.toml`
    - Shuts down hotspot
    - Connects to configured Wi-Fi
    - Enters normal operation

### Normal operation
- On LAN, the dashboard remains accessible via `http://aurapack.local` (mDNS).
- The assistant connects to the central MQTT broker (usually hosted on the LLM server).
- Whisper.cpp and Piper run locally, invoked as persistent subprocesses via IPC or CLI.
- The assistant is stateless and resilient:
  - systemd `Restart=always`
  - health pings via `/system/health`
  - STT/TTS subprocesses auto-restart if they fail.

### Security
- The configuration dashboard binds to `127.0.0.1` by default; LAN access must be explicitly enabled.
- MQTT credentials are unique per node and stored locally.
- No raw audio retained unless explicitly enabled.

## Result
- Fully reproducible build and deployment.
- Nice user experience (plug-and-configure through hotspot).
- systemd integration.
- Easy future extension to support multiple assistants on the same network.

