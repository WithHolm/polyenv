package plugin

import "fmt"

type StdOutWriter struct{}

func (e *StdOutWriter) Write(data []byte) error {
	fmt.Println(string(data))
	return nil
}

func (e *StdOutWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"*"}, []string{}
}
