# ADR-0002 - Intent-to-Action Pipeline and Device Control 

## Status
Proposed — 2025-11-01

## Context
The **assistant** node can already capture audio, perform STT (Whisper), and output TTS (Piper).  
The missing piece is a deterministic pipeline that translates **spoken intents** into **concrete, observable device actions**, without coupling the LLM to specific hardware protocols.

At this stage, Aurapack’s first implementation phase will focus on **Arduino** and **ESP32** boards communicating directly over **MQTT**. These microcontrollers will act as low-level actuators and sensors.  
Future extensions (Zigbee2MQTT, ESPHome, Modbus, HTTP bridges, etc.) will reuse the same orchestration layer and schema but are not part of the initial MVP.

Goals and constraints:
- Keep **LLM semantic and hardware-agnostic** (it only emits normalized intents, never direct device commands).
- Use **MQTT** as the backbone for intent, state, and action routing.
- Ensure **idempotency**, **observability**, and **debuggability** (every step traceable with correlation IDs).
- Make every component **restart-safe**, **message-driven**, and decoupled.

## Decision

### 1) High-level architecture

```
Voice → STT → LLM (intent) → Orchestrator (Raspberry) → Device Adapter(s) → Hardware
   ↑                                                                ↓
   └───────────── Confirmation / State Events via MQTT ─────────────┘
```

Roles:
- **Assistant**: captures audio, sends text to LLM, speaks confirmations or errors.
- **LLM Service**: converts natural language into normalized **intents** (JSON schema).
- **Orchestrator** (Raspberry Pi): validates intents, authorizes, routes to device adapters, tracks execution, publishes acks/results.
- **Device Adapters**: translate generic device commands into actual control messages (initially Arduino/ESP32 over MQTT).
- **Devices**: endpoint microcontrollers exposing `.../set` and `.../state` topics.

### 2) MQTT topics (core convention)

- `aurapack/assistant/<node_id>/nlu/request`
- `aurapack/assistant/<node_id>/nlu/response`
- `aurapack/intent/commands`
- `aurapack/device/<device_id>/set`
- `aurapack/device/<device_id>/state`
- `aurapack/exec/<exec_id>/ack`
- `aurapack/exec/<exec_id>/result`
- `aurapack/registry/announce`

### 3) Intent JSON contract (LLM output)

```json
{
  "schema": "aurapack.intent/1.0",
  "intent_id": "3d2a5a3c-3b0a-4d3c-9b9d-0a6c2f7b0f41",
  "timestamp": "2025-11-01T14:58:00Z",
  "actor": { "user_id": "adrian", "session_id": "demo" },
  "action": {
    "verb": "turn_on",
    "target": {
      "kind": "light",
      "selector": { "name": "living_room.ceiling", "room": "living_room" }
    },
    "params": { "brightness": 80 },
    "constraints": { "require_confirmation": true }
  },
  "meta": {
    "utterance": "Turn on the living room light to eighty percent.",
    "confidence": 0.86
  }
}
```

Notes:
- The **LLM never specifies GPIO or hardware IDs**.
- The **Orchestrator** resolves selectors to devices via a retained registry.
- `intent_id` is used for **idempotency** and deduplication.
- Actions are normalized to verbs and targets (`turn_on`, `light.living_room.ceiling`).

### 4) Orchestrator responsibilities (Raspberry)

1. Resolve the target (`selector → device_id` from retained registry).
2. Authorize (`actor` against policy rules).
3. Normalize units and validate parameters.
4. Emit an `ack` and forward the action to the appropriate device topic.
5. Track completion via device `.../state` updates.
6. Publish `result` or `error` events.
7. Bridge a **speakable summary** back to the Assistant via MQTT for TTS.

### 5) Device-level interaction (Arduino/ESP32 MVP)

Each board runs a lightweight firmware exposing MQTT topics:

**Example (Arduino with PubSubClient):**
```cpp
client.publish("aurapack/device/living_room_light/state", "{"on":false}");

void callback(char* topic, byte* payload, unsigned int length) {
  if (strcmp(topic, "aurapack/device/living_room_light/set") == 0) {
    bool turn_on = strstr((char*)payload, ""on":true");
    digitalWrite(RELAY_PIN, turn_on ? HIGH : LOW);
    client.publish("aurapack/device/living_room_light/state", turn_on ? "{"on":true}" : "{"on":false}");
  }
}
```

### 6) Error taxonomy
- `unauthorized`
- `invalid_intent`
- `ambiguous_target`
- `target_not_found`
- `device_unreachable`
- `driver_error`
- `deadline_exceeded`
- `conflict`
- `safety_blocked`

### 7) Speakable responses (UX contract)

| Result | Spoken message |
|---------|----------------|
| success | “Turned on the living room ceiling to 80 percent.” |
| ambiguous | “There are two lights called ceiling. Which one?” |
| device_unreachable | “The light didn’t respond.” |
| unauthorized | “You’re not allowed to control that device.” |

### 8) Idempotency and QoS
- `intent_id` deduplication window (30s default)
- Commands and results use QoS 1 (at-least-once)
- State topics retained with QoS 0
- Device registry retained for quick recovery

### 9) Security
- Unique MQTT credentials per node.
- Topic ACLs prevent assistants from publishing to device topics directly.
- Orchestrator enforces authorization policies.

### 10) Observability
- Structured logs with `intent_id`, `exec_id`, `device_id`.
- Prometheus metrics: intent count, success ratio, latency histograms.
- Distributed tracing with correlation propagation through MQTT.

### 11) Example end-to-end flow

```
User → Assistant: “Turn on the living room light.”
Assistant → LLM: publish /nlu/request
LLM → Orchestrator: publish /intent/commands (JSON intent)
Orchestrator → ack
Orchestrator → /device/living_room_light/set → {"on":true}
Device → /device/living_room_light/state → {"on":true}
Orchestrator → /exec/result → {"status":"success"}
Orchestrator → /assistant/output/text → "Turned on the light."
Assistant → Piper → speaks confirmation
```

### 12) Future extensions
While the initial focus is Arduino and ESP32, the schema supports plug‑in adapters for:
- Zigbee2MQTT
- ESPHome
- HTTP/REST bridges
- Modbus/TCP industrial IO
Each adapter follows the same `Resolve / Plan / Execute / Observe` pattern and uses the same event schema.

## Result
- Cohesive semantic‑to‑action workflow, MQTT‑based, modular, and observable.
- Clear separation between **intent resolution** and **device execution**.
- Extensible to future protocols without retraining or modifying the LLM.
- Reliable end‑to‑end confirmation from speech → intent → action → voice.
