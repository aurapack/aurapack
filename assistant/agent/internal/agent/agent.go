package agent

import (
	"fmt"
	"log"
)

// STT converts speech to text
type STT interface {
	Listen() (string, error)
}

// TTS converts text to speech
type TTS interface {
	Speak(text string) error
}

// Agent wires STT, TTS, and MQTT loop.
type Agent struct {
	stt  STT
	tts  TTS
	mqtt *MQTT
}

// New builds an Agent.
func New() *Agent {
	m, err := NewMQTT(MQTTBroker)
	if err != nil {
		log.Fatalf("[Agent] MQTT connect error: %v", err)
	}
	return &Agent{stt: &FakeSTT{}, tts: &FakeTTS{}, mqtt: m}
}

// NewWith lets you inject concrete STT/TTS implementations.
func NewWith(stt STT, tts TTS, brokerURL string) *Agent {
	m, err := NewMQTT(brokerURL)
	if err != nil {
		log.Fatalf("[Agent] MQTT connect error: %v", err)
	}
	return &Agent{stt: stt, tts: tts, mqtt: m}
}

// Run subscribes to input topic and publishes responses on output topic.
func (a *Agent) Run() {
	fmt.Println("[Agent] Connected to MQTT broker at", a.mqtt.brokerURL)
	a.mqtt.Subscribe(InTopic, a.handleMessage)
	select {}
}

func (a *Agent) handleMessage(payload []byte) {
	msg := string(payload)
	fmt.Println("[Agent] Received:", msg)
	// NOTE: For now, trivial response. Later: intent parsing, LLM, devices.
	resp := "Echo: " + msg
	_ = a.tts.Speak(resp)
	a.mqtt.Publish(OutTopic, []byte(resp))
}
