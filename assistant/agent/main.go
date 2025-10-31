package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

const (
	host     = "tcp://localhost:1883"
	inTopic  = "/assistant/input/text"
	outTopic = "/assistant/output/text"
)

func handleInput(c mqtt.Client, m mqtt.Message) {
	msg := string(m.Payload())
	fmt.Println("Received:", msg)
	c.Publish(outTopic, 1, false, "Echo: "+msg)
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker(host).SetClientID("assistant-dev")
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	fmt.Println("Connected to MQTT broker at", host)
	if token := client.Subscribe(inTopic, 1, handleInput); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	select {} // keep running
}
