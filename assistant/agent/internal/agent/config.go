package agent

import (
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

type MQTTConfig struct {
	Broker   string `toml:"broker"`
	InTopic  string `toml:"in_topic"`
	OutTopic string `toml:"out_topic"`
}

type STTConfig struct {
	Bin       string   `toml:"bin"`
	Model     string   `toml:"model"`
	RecordCmd []string `toml:"record_cmd"`
}

type TTSConfig struct {
	Bin   string `toml:"bin"`
	Voice string `toml:"voice"`
}

type Config struct {
	MQTT MQTTConfig `toml:"mqtt"`
	STT  STTConfig  `toml:"stt"`
	TTS  TTSConfig  `toml:"tts"`
}

var (
	cfg     *Config
	cfgOnce sync.Once
)

// LoadConfig reads assistant.toml from ASSISTANT_CFG (file path) or defaults to /opt/aurapack/assistant/cfg/assistant.toml
func LoadConfig() *Config {
	cfgOnce.Do(func() {
		path := os.Getenv("ASSISTANT_CFG")
		if path == "" {
			path = "/opt/aurapack/assistant/cfg/assistant.toml"
		}
		var c Config
		if _, err := toml.DecodeFile(path, &c); err != nil {
			c = Config{
				MQTT: MQTTConfig{
					Broker:   "tcp://localhost:1883",
					InTopic:  "/assistant/input/text",
					OutTopic: "/assistant/output/text",
				},
				STT: STTConfig{
					Bin:       "/opt/aurapack/assistant/bin/whisper",
					Model:     "/opt/aurapack/assistant/models/ggml-base.en.bin",
					RecordCmd: []string{"arecord", "-q", "-f", "S16_LE", "-r", "16000", "-c1", "-d", "5", "/tmp/aurapack_in.wav"},
				},
				TTS: TTSConfig{
					Bin:   "/opt/aurapack/assistant/bin/piper",
					Voice: "/opt/aurapack/assistant/models/en_US-amy-medium.onnx",
				},
			}
		}
		cfg = &c
	})
	return cfg
}
