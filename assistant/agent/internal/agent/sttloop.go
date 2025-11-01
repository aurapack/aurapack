package agent

import (
    "encoding/json"
    "fmt"
    "time"
)

type sttPayload struct {
    Text      string `json:"text"`
    SessionID string `json:"session_id"`
}

// STTLoop continuously records and publishes transcriptions into the input topic.
func STTLoop(stt STT, mqtt *MQTT, inTopic string) {
    for {
        text, err := stt.Listen()
        if err != nil {
            fmt.Println("[STTLoop] error:", err)
            time.Sleep(1 * time.Second)
            continue
        }
        if text == "" {
            continue
        }
        payload, _ := json.Marshal(sttPayload{Text: text, SessionID: "local"})
        mqtt.Publish(inTopic, payload)
    }
}
