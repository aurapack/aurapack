package agent

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTT struct {
	client    mqtt.Client
	brokerURL string
	once      sync.Once
}

func NewMQTT(brokerURL string) (*MQTT, error) {
	opts := mqtt.NewClientOptions().AddBroker(brokerURL).SetClientID("assistant-agent")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &MQTT{client: client, brokerURL: brokerURL}, nil
}

func (m *MQTT) Subscribe(topic string, cb func([]byte)) {
	h := func(c mqtt.Client, msg mqtt.Message) { cb(msg.Payload()) }
	token := m.client.Subscribe(topic, 1, h)
	token.Wait()
}

func (m *MQTT) Publish(topic string, payload []byte) {
	token := m.client.Publish(topic, 1, false, payload)
	token.Wait()
	fmt.Printf("[MQTT] Published to %s\n", topic)
}
