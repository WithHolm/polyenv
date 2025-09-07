package plugin

import (
	"os"
)

type StdOutWriter struct{}

func (e *StdOutWriter) Write(data []byte) error {
	_, err := os.Stdout.Write(data)
	if err != nil {
		return err
	}
	// fmt.Println(string(data))
	return nil
}

func (e *StdOutWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"*"}, []string{}
}
