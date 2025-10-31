package agent

import (
	"bufio"
	"fmt"
	"os"
)

// FakeSTT reads a line from stdin to simulate speech input.
type FakeSTT struct{}

func (f *FakeSTT) Listen() (string, error) {
	fmt.Print("[FakeSTT] Say something (type text): ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}
