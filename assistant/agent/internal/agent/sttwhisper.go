package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type WhisperSTT struct {
	Bin       string
	Model     string
	RecordCmd []string
}

func (w *WhisperSTT) Listen() (string, error) {
	record := exec.Command(w.RecordCmd[0], w.RecordCmd[1:]...)
	if err := record.Run(); err != nil {
		return "", fmt.Errorf("recording failed: %v", err)
	}

	cmd := exec.Command(w.Bin, "--model", w.Model, "/tmp/aurapack_in.wav")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("whisper failed: %v", err)
	}

	lines := strings.Split(out.String(), "\n")
	re := regexp.MustCompile(`^\[[^\]]+\]\s*(.*)$`)
	var buf []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if m := re.FindStringSubmatch(line); len(m) == 2 {
			if text := strings.TrimSpace(m[1]); text != "" {
				buf = append([]string{text}, buf...)
			}
			continue
		}
		buf = append([]string{line}, buf...)
	}
	if len(buf) > 0 {
		return strings.Join(buf, " "), nil
	}

	return strings.TrimSpace(out.String()), nil
}
