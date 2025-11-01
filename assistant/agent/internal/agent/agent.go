package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type STT interface {
	Listen() (string, error)
}

type TTS interface {
	Speak(text string) error
}

type Agent struct {
	stt  STT
	tts  TTS
	mqtt *MQTT
	cfg  *Config
}

func New() *Agent {
	cfg := LoadConfig()

	stt := &WhisperSTT{
		Bin:       cfg.STT.Bin,
		Model:     cfg.STT.Model,
		RecordCmd: cfg.STT.RecordCmd,
	}

	tts := &PiperTTS{
		Bin:   cfg.TTS.Bin,
		Voice: cfg.TTS.Voice,
	}

	m, err := NewMQTT(cfg.MQTT.Broker)
	if err != nil {
		log.Fatalf("[Agent] MQTT connect error: %v", err)
	}

	return &Agent{stt: stt, tts: tts, mqtt: m, cfg: cfg}
}

func (a *Agent) Run() {
	fmt.Println("[Agent] Connected to MQTT broker at", a.mqtt.brokerURL)
	a.mqtt.Subscribe(a.cfg.MQTT.InTopic, a.handleMessage)
	go STTLoop(a.stt, a.mqtt, a.cfg.MQTT.InTopic)
	select {}
}

type inMsg struct {
	Text      string `json:"text"`
	SessionID string `json:"session_id"`
}

func (a *Agent) handleMessage(payload []byte) {
	var msg inMsg
	if err := json.Unmarshal(payload, &msg); err != nil {
		fmt.Println("[Agent] invalid JSON on input topic:", err)
		return
	}
	if msg.Text == "" {
		return
	}
	fmt.Println("[Agent] Received:", msg.Text)
	resp := "Echo: " + msg.Text
	if err := a.tts.Speak(resp); err != nil {
		fmt.Println("[Agent] TTS error:", err)
	}

	a.mqtt.Publish(a.cfg.MQTT.OutTopic, []byte(resp))
	time.Sleep(10 * time.Millisecond)
}
