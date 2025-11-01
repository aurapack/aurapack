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

	lines := strings.Split(out.String(), "")
	re := regexp.MustCompile(`^\[[^\]]+\]\s*(.*)$`)
	for i := len(lines) - 1; i >= 0; i-- {
		m := re.FindStringSubmatch(strings.TrimSpace(lines[i]))
		if len(m) == 2 {
			return strings.TrimSpace(m[1]), nil
		}
	}

	return strings.TrimSpace(out.String()), nil
}
