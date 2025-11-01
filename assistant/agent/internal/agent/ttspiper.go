package agent

import (
    "bytes"
    "fmt"
    "os/exec"
)

type PiperTTS struct {
    Bin   string
    Voice string
}

func (p *PiperTTS) Speak(text string) error {
    cmd := exec.Command(p.Bin, "--model", p.Voice, "--output_file", "/tmp/aurapack_out.wav")
    cmd.Stdin = bytes.NewBufferString(text)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("piper failed: %v", err)
    }
    play := exec.Command("aplay", "-q", "/tmp/aurapack_out.wav")
    return play.Run()
}
