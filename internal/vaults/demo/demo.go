package demo

import "github.com/charmbracelet/huh"

type Client struct {
}

func (t *Client) DisplayName() string {
	return "Temp Vault"
}

func (t *Client) Push(name string, value string) error {
	return nil
}

func (t *Client) Pull() (map[string]string, error) {
	return map[string]string{}, nil
}

func (t *Client) List() ([]string, error) {
	return []string{}, nil
}

func (t *Client) Warmup() error {
	return nil
}

func (t *Client) ValidateConfig(options map[string]string) error {
	return nil
}

func (t *Client) SetOptions(map[string]string) error {
	return nil
}

func (t *Client) GetOptions() map[string]string {
	return map[string]string{}
}

func (t *Client) Opsie() error {
	return nil
}

func (t *Client) Flush(key []string) error {
	return nil
}

func (t *Client) NewWizardWarmup() error {
	return nil
}

func (t *Client) NewWizardNext() *huh.Form {
	return nil
}

func (t *Client) NewWizardComplete() map[string]string {
	return map[string]string{}
}

func (t *Client) UpdateWizardWarmup(map[string]string) error {
	return nil
}

func (t *Client) UpdateWizardNext() *huh.Form {
	return nil
}

func (t *Client) UpdateWizardComplete() map[string]string {
	return map[string]string{}
}

// func

// func (t *temp) Next() *huh.Form {
// 	re
// }

// func (t *temp) Complete() map[string]string {
// 	panic("implement me")
// }
