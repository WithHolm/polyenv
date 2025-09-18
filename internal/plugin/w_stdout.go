package plugin

import (
	"os"
)

type StdOutWriter struct{}

func (e *StdOutWriter) Name() string {
	return "stdout"
}

func (e *StdOutWriter) Write(data []byte) error {
	_, err := os.Stdout.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (e *StdOutWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"stats", "*"}, []string{}
}
