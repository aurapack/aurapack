package agent

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PiperTTS struct {
	Bin   string
	Voice string
	Args  []string
	Env   map[string]string
}

func (p *PiperTTS) Speak(text string) error {
	args := append([]string{"--model", p.Voice}, p.Args...)
	outputFile := "/tmp/aurapack_out.wav"
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--output_file" {
			outputFile = args[i+1]
			break
		}
	}

	cmd := exec.Command(p.Bin, args...)
	cmd.Stdin = bytes.NewBufferString(text)
	cmd.Env = mergeEnv(os.Environ(), p.Env)

	if out, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("piper failed: %v: %s (cmd: %s)", err, msg, strings.Join(cmd.Args, " "))
		}
		return fmt.Errorf("piper failed: %w (cmd: %s)", err, strings.Join(cmd.Args, " "))
	}

	play := exec.Command("aplay", "-q", outputFile)
	if out, err := play.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("aplay failed: %v: %s (cmd: %s)", err, msg, strings.Join(play.Args, " "))
		}
		return fmt.Errorf("aplay failed: %w (cmd: %s)", err, strings.Join(play.Args, " "))
	}
	return nil
}

func mergeEnv(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	filtered := make([]string, 0, len(base))
	for _, kv := range base {
		sep := strings.IndexByte(kv, '=')
		if sep == -1 {
			filtered = append(filtered, kv)
			continue
		}
		key := kv[:sep]
		if _, ok := overrides[key]; ok {
			continue
		}
		filtered = append(filtered, kv)
	}
	for k, v := range overrides {
		filtered = append(filtered, fmt.Sprintf("%s=%s", k, v))
	}
	return filtered
}
