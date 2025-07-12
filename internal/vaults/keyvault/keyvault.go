/*
notify:@withholm
*/
package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/tools"
)

type Client struct {
	//name of the keyvault
	name string

	//tenant of the keyvault
	tenant string

	//name of the tag that contains the env key
	tag string

	//uri of the keyvault
	uri string

	//include keys and certificates
	// includeCert bool

	autoUppercase bool

	replaceHyphen bool

	appendExpiration string

	ignoreContentType []string

	keys []string

	client *azsecrets.Client

	wiz Wizard
}

var ctx context.Context

func Init() {
	os.Setenv("AZURE_SDK_GO_LOGGING", "all")
	azlog.SetListener(func(cls azlog.Event, msg string) {
		slog.Debug(msg)
	})
	azlog.SetEvents(azlog.EventRequest, azlog.EventResponse)
}

// validate that client implemts the vault interface -> done at vaults.go to avoid circular dependency

// returns the display name of the vault
func (cli *Client) DisplayName() string {
	return "Azure Key Vault"
}

//region new wiz

func (cli *Client) WizardWarmup(m map[string]string) error {
	cli.wiz = newWizard(m)
	cli.wiz.Warmup()
	return nil
}

func (cli *Client) WizardNext() *huh.Form {
	return cli.wiz.Next()
}

func (cli *Client) WizardComplete() map[string]string {
	return cli.wiz.Complete()
}

// set attributes for the client. used by repository init
func (cli *Client) SetOptions(options map[string]string) error {
	slog.Debug("setting options", "options", options)

	cli.tag = options["ENV_NAME_TAG"]
	cli.uri = options["URI"]
	cli.name = options["NAME"]
	cli.tenant = options["TENANT"]
	cli.autoUppercase = options["AUTO_UPPERCASE"] == "true"
	cli.replaceHyphen = options["REPLACE_HYPHEN"] == "true"
	cli.appendExpiration = options["APPEND_EXPIRATION"]
	cli.ignoreContentType = strings.Split(options["IGNORE_CONTENT_TYPES"], ",")

	if cli.appendExpiration != "" {
		_, err := time.Parse(time.RFC3339, cli.appendExpiration)
		if err != nil {
			return fmt.Errorf("failed to parse append expiration: %s", err)
		}
	}

	if len(cli.ignoreContentType) != 0 && len(cli.keys) != 0 {
		return fmt.Errorf("cannot set both keys and ignore content types")
	}

	if cli.tag == "" {
		return fmt.Errorf("env name tag cannot be empty")
	}
	if cli.uri == "" {
		return fmt.Errorf("uri for keyvault cannot be empty")
	}
	if cli.name == "" {
		return fmt.Errorf("name of keyvault cannot be empty")
	}
	if cli.tenant == "" {
		return fmt.Errorf("tenant for keyvault cannot be empty")
	}

	return nil
}

// Converts a string to keyvault name
func ConvertToKeyvaultName(value string) string {
	value = strings.ToLower(value)
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	result := re.ReplaceAllStringFunc(value, func(r string) string {

		return "-"
	})
	return result
}

// Push pushes a single secret to keyvault
func (cli *Client) Push(name string, value string) error {
	contentType := "text/plain"
	secretparam := azsecrets.SetSecretParameters{
		Value:       &value,
		ContentType: &contentType,
		Tags: map[string]*string{
			cli.tag: &name,
		},
	}
	sName := ConvertToKeyvaultName(name)
	_, err := cli.client.SetSecret(context.Background(), sName, secretparam, nil)
	if err != nil {
		return fmt.Errorf("failed to push secret: %s", err)
	}

	return nil
}

// Pull  all secrets from keyvault
func (cli *Client) Pull() (map[string]string, error) {
	out := make(map[string]string)

	list, err := cli.List()
	if err != nil {
		return nil, err
	}
	for _, secret := range list {
		val, err := cli.client.GetSecret(context.Background(), secret, "", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret %s: %s", secret, err)
		}
		// try to get val from tags, else just use secret name
		var n string
		if val.Secret.Tags[cli.tag] == nil {
			slog.With("secret", secret).Debug("no env key found in tags, using secret name")
			n = secret
		} else {
			n = *val.Secret.Tags[cli.tag]
		}
		out[n] = *val.Secret.Value
	}

	return out, nil
}

// List all secrets
func (cli *Client) List() (out []string, err error) {
	// chn := make(chan *azsecrets.SecretProperties)
	list, err := listSecrets(cli.client)
	if err != nil {
		return nil, err
	}
	for _, secret := range list {
		// if !cli.includeCert && *secret.ContentType == "application/x-pkcs12" {
		// 	continue
		// }
		out = append(out, secret.ID.Name())
	}

	return out, nil
}

// func so i can run it with wizard aswell..
func listSecrets(client *azsecrets.Client) (out []*azsecrets.SecretProperties, err error) {
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %s", err)
		}
		for _, secret := range page.Value {
			out = append(out, secret)
		}
	}
	return out, nil
}

// region flush
// Flush flushes a single secret from keyvault
func (cli *Client) Flush(key []string) error {
	list, err := cli.List()
	if err != nil {
		return err
	}
	for _, name := range list {
		if slices.Contains(key, name) {
			_, err := cli.client.DeleteSecret(context.Background(), name, nil)
			if err != nil {
				return fmt.Errorf("failed to delete secret: %s", err)
			}
		}
	}
	return nil
}

// endregion flush
func (cli *Client) Opsie() error {
	return fmt.Errorf("not implemented yet")
}

func checkAzCliInstalled() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("az cli not installed. please install it and try again")
	}
	return nil
}

func (cli *Client) Warmup() error {
	err := checkAzCliInstalled()
	if err != nil {
		return err
	}

	if cli.tenant == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: cli.tenant,
	})
	if err != nil {
		return err
	}

	newCli, err := azsecrets.NewClient(cli.uri, cred, nil)
	if err != nil {
		return err
	}
	cli.client = newCli

	return nil
}

func (cli *Client) ValidateConfig(options map[string]string) error {
	// j, e := json.MarshalIndent(options, "", "  ")
	// if e != nil {
	// 	return fmt.Errorf("failed to marshal options: %s", e)
	// }
	// slog.Info("options", "options", string(j))

	if options["VAULT_TYPE"] != "keyvault" {
		return fmt.Errorf("invalid vault type: %s. expecting keyvault", options["VAULT_TYPE"])
	}

	if options["TENANT"] == "" {
		return fmt.Errorf("TENANT cannot be empty")
	}

	//TODO: add url validation?
	if options["URI"] == "" {
		return fmt.Errorf("URI cannot be empty")
	}

	//actually.. no.. you can have keys set to empty and ignore content types set to empty (ie pull all)
	// if options["KEYS"] == "" && options["IGNORE_CONTENT_TYPES"] == "" {
	// 	return fmt.Errorf("either KEYS or IGNORE_CONTENT_TYPES must be set")
	// }

	//validate if its a json array
	for _, v := range []string{"KEYS", "IGNORE_CONTENT_TYPES"} {
		if options[v] == "" {
			continue
		}
		var keys []string
		err := json.Unmarshal([]byte(options[v]), &keys)
		if err != nil {
			return fmt.Errorf("failed to convert %s to json array: %s", v, err)
		}
	}

	if options["APPEND_EXPIRATION"] == "" {
		return fmt.Errorf("APPEND_EXPIRATION cannot be empty")
	} else {
		err := tools.ValidateIsoDate(options["APPEND_EXPIRATION"])
		if err != nil {
			return fmt.Errorf("APPEND_EXPIRATION validation error: %s", err)
		}
	}

	//validate boolean values
	for _, v := range []string{"REPLACE_HYPHEN", "AUTO_UPPERCASE"} {
		// slog.Info("validating", "key", v, "value", options[v])
		if options[v] == "" {
			return fmt.Errorf("%s must be defined", v)
		} else if !slices.Contains([]string{"true", "false"}, strings.ToLower(options[v])) {
			return fmt.Errorf("%s must be boolean", v)
		}
	}

	return nil
}
