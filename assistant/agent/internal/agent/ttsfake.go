package agent

import "fmt"

// FakeTTS prints the text as if it were spoken.
type FakeTTS struct{}

func (f *FakeTTS) Speak(text string) error {
	fmt.Println("[FakeTTS] Speaking:", text)
	return nil
}
