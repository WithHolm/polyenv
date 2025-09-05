package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/sosodev/duration"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
)

type OtsWriter struct {
	concealOpts *otsOptions
}

type otsOptions struct {
	Url           string
	ConcealMaxTTL *duration.Duration
}

type otsConcealSecret struct {
	Secret         string  `json:"secret"`
	RecipientEmail string  `json:"recipient,omitempty"`
	PassPrase      string  `json:"passphrase"`
	TTL            float64 `json:"ttl"`
}

type otsConcealResponse struct {
	Success bool   `json:"success"`
	Shrimp  string `json:"shrimp"`
	Record  struct {
		Metadata struct {
			Identifier  string  `json:"identifier"`
			SecretTTL   float64 `json:"secret_ttl"`
			MetadataTTL float64 `json:"metadata_ttl"`
		} `json:"metadata"`
		Secret struct {
			Identifier string `json:"identifier"`
			Secret     string `json:"secret"`
		} `json:"secret"`
	} `json:"record"`
}

func (r *otsConcealResponse) Expires() (time.Time, error) {
	duration := time.Duration(r.Record.Metadata.SecretTTL) * time.Second
	return time.Now().Add(duration), nil
}

func (e *OtsWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"pick", "dotenv", "*"}, []string{}
}

func (e *OtsWriter) GetTTL(s string) (*duration.Duration, error) {
	s = strings.ToUpper(s)

	if s == "" {
		s = e.concealOpts.ConcealMaxTTL.String()
	}

	if !strings.HasPrefix(s, "P") {
		days := 0
		fmt.Sscanf(s, "%d", &days)
		slog.Debug("got days from input, converting to duration", "days", days)
		s = "P" + s + "D"
	}

	duration, err := duration.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ISO duration '%s': %w", s, err)
	}

	return duration, nil
}

func (e *OtsWriter) GetOptions() (*otsOptions, error) {
	d, err := e.GetTTL("P7D")
	if err != nil {
		return nil, fmt.Errorf("failed to get TTL: %w", err)
	}
	return &otsOptions{
		Url:           "eu.onetimesecret.com",
		ConcealMaxTTL: d,
	}, nil
}

// ses options by https://docs.onetimesecret.com/en/rest-api/ to deliver form
func (e *OtsWriter) Write(data []byte) error {
	var err error
	e.concealOpts, err = e.GetOptions()
	if err != nil {
		return fmt.Errorf("failed to get options: %w", err)
	}

	secret := otsConcealSecret{
		Secret:    string(data),
		TTL:       e.concealOpts.ConcealMaxTTL.ToTimeDuration().Seconds(),
		PassPrase: "",
	}
	var ttl string
	form := huh.NewForm(
		huh.NewGroup(
			// huh.NewInput[string]().
			// 	Title("Recipient email").
			// 	Description("The email address of the recipient (enter to leave empty)").
			// 	Required(true).
			// 	Value(&secret.RecipientEmail),
			huh.NewInput().
				Title("Passphrase").
				Description("The passphrase to use for encryption").
				Placeholder("Enter to leave empty").
				Value(&secret.PassPrase),
			huh.NewInput().
				CharLimit(4).
				Title("TTL").
				Description("How long the secret should be valid for? either number of days or iso duration (max P7D)").
				Placeholder(e.concealOpts.ConcealMaxTTL.String()).
				Value(&ttl).
				Validate(func(s string) error {
					dur, err := e.GetTTL(s)
					if err != nil {
						return err
					}
					if dur.Days > 7 {
						return fmt.Errorf("max duration is 7 days (P7D)")
					}
					return nil
				}),
		),
	)
	tui.RunHuh(form)

	var usingTTL *duration.Duration
	usingTTL, err = e.GetTTL(ttl)

	secret.TTL = float64(usingTTL.ToTimeDuration().Seconds())

	body := map[string]any{
		"secret": secret,
	}
	resp := otsConcealResponse{}

	httpClient := tools.NewPolyenvHttpClient()
	err = httpClient.Post(context.Background(), "https://eu.onetimesecret.com/api/v2/secret/conceal", body, &resp)
	if err != nil {
		return fmt.Errorf("failed to post to OneTimeSecret: %w", err)
	}
	respjson, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Debug("response", "json", string(respjson))
	slog.Info("response", "success", resp.Success, "Shrimp", resp.Shrimp)
	exp, err := resp.Expires()
	if err != nil {
		return err
	}
	out := []string{
		"generated via 'ots' plugin for polyenv (github.com/withholm/polyenv)",
		fmt.Sprintf("url: https://eu.onetimesecret.com/secret/%s", resp.Record.Secret.Identifier),
		fmt.Sprintf("it will expire in %s", exp.Format("2006-01-02 15:04:05z")),
		"if you already have polyenv, you can run 'polyenv !{yourenv} import {link or slug}' to import the secret to your env",
	}
	fmt.Println(strings.Join(out, "\n"))
	return nil
}
