package keyvault

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

type (
	Wizard struct {
		Tenant        string
		Tenants       []string
		Subscription  string
		Subscriptions []string
		Uri           string
		Name          string
		state         int
	}

	GraphQueryItem struct {
		Name           string
		ResourceGroup  string
		SubscriptionId string
		TenantId       string
		VaultUri       string
		Location       string
	}
)

// region new wiz
func (cli *Client) WizWarmup(m map[string]any) error {
	cli.wiz = Wizard{
		Tenant:       "",
		Subscription: "",
		Uri:          "",
		Name:         "",
		state:        0,
	}

	if m["tenant"] != nil {
		var e error
		cli.wiz.Tenant, e = GetTenant(m["tenant"].(string))
		if e != nil {
			return e
		}
	}

	if m["subscription"] != nil {
		cli.wiz.Subscription = m["subscription"].(string)
	}
	if m["sub"] != nil {
		cli.wiz.Subscription = m["sub"].(string)
	}

	// if m["uri"] != nil {
	// 	cli.wiz.Uri = m["uri"].(string)
	// }

	// if m["name"] != nil {
	// 	cli.wiz.Name = m["name"].(string)
	// }

	for k := range m {
		if k == "tenant" || k == "subscription" || k == "sub" || k == "uri" || k == "name" {
			continue
		}
		v := m[k]
		slog.Warn("unknown key for keyvault wizard", "key", k, "value", v)
	}

	err := checkAzCliInstalled()
	if err != nil {
		return err
	}

	return nil
}

func (cli *Client) WizNext() (*huh.Form, error) {
	// automatically increment the formGroup
	defer func() { cli.wiz.state++ }()
	switch cli.wiz.state {
	case 0: //select tenant
		if cli.wiz.Tenant != "" {
			slog.Debug("skipping tenant pick", "tenant", cli.wiz.Tenant)
			cli.wiz.state++
			return cli.WizNext()
		}

		// v, e := getTenants()
		// if e != nil {
		// 	//TODO: Handle error instead of OS.Exit
		// 	slog.Error("failed to get tenants: " + e.Error())
		// 	os.Exit(1)
		// }

		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a tenant").
					OptionsFunc(func() []huh.Option[string] {
						v, e := getTenants()
						if e != nil {
							panic(fmt.Errorf("failed to get tenants: %w", e))
						}
						// 	//TODO: Handle error instead of OS.Exit
						// 	slog.Error("failed to get tenants: " + e.Error())
						// 	os.Exit(1)
						// }
						ret := make([]huh.Option[string], 0)
						for _, tenant := range v {
							opt := huh.NewOption(*tenant.DisplayName, *tenant.TenantID)
							ret = append(ret, opt)
						}
						return ret
					}, nil).
					Value(&cli.wiz.Tenant)),
		), nil
	case 1: //select subscription and vault
		fields := make([]huh.Field, 0)

		if cli.wiz.Subscription == "" {
			slog.Debug("Showing subscriptions")
			fields = append(fields, huh.NewSelect[string]().
				Title("Select Subscription").
				OptionsFunc(func() (ret []huh.Option[string]) {
					subs, err := getSubscriptions(cli.wiz.Tenant)
					if err != nil {
						slog.Error("failed to get subscriptions: " + err.Error())
						os.Exit(1)
					} else if len(subs) == 0 {
						slog.Error("no subscriptions found. please auth to tenant: az login --tenant " + cli.wiz.Tenant)
						os.Exit(1)
					}
					for _, sub := range subs {
						opt := huh.NewOption(*sub.DisplayName, *sub.SubscriptionID)
						ret = append(ret, opt)
					}
					return ret
				}, cli.wiz.Tenant).Value(&cli.wiz.Subscription),
			)
		} else { // if subscription is defined
			subscriptions, e := getSubscriptions(cli.wiz.Tenant)
			if e != nil {
				slog.Error("failed to get subscriptions: " + e.Error())
				os.Exit(1)
			}
			if len(subscriptions) == 0 {
				slog.Error("no subscriptions found. please auth to tenant: az login --tenant " + cli.wiz.Tenant)
				os.Exit(1)
			}
			regexGuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
			isGuid := regexGuid.MatchString(cli.wiz.Subscription)
			found := false
			for _, sub := range subscriptions {
				slog.Debug("checking subscription", "id", *sub.SubscriptionID, "name", *sub.DisplayName)
				//look for guid
				if isGuid && strings.EqualFold(*sub.SubscriptionID, cli.wiz.Subscription) {
					slog.Debug("found subscription", "id", *sub.SubscriptionID, "name", *sub.DisplayName)
					cli.wiz.Subscription = *sub.SubscriptionID
					found = true
					break
				}
				//look for display name
				if !isGuid && strings.EqualFold(*sub.DisplayName, cli.wiz.Subscription) {
					slog.Debug("found subscription", "id", *sub.SubscriptionID, "name", *sub.DisplayName)
					cli.wiz.Subscription = *sub.SubscriptionID
					found = true
					break
				}
			}
			if !found {
				slog.Error("subscription not found", "value", cli.wiz.Subscription, "is guid", isGuid)
				os.Exit(1)
			}

			slog.Debug("Skipping Subscription", "defined", cli.wiz.Subscription)
		}

		if cli.wiz.Name == "" {
			fields = append(fields, huh.NewSelect[string]().
				Title("Select Vault").
				OptionsFunc(func() (ret []huh.Option[string]) {
					vaults, err := getKeyvaults(cli.wiz.Subscription, cli.wiz.Tenant)
					if err != nil {
						slog.Error("failed to get vaults: " + err.Error())
						os.Exit(1)
					}
					for _, vault := range vaults {
						opt := huh.NewOption(vault.Name, vault.VaultUri)
						ret = append(ret, opt)
					}
					return ret
				}, &cli.wiz.Subscription).Value(&cli.wiz.Uri),
			)
		}
		//TODO: else check if vault exists

		return huh.NewForm(
			huh.NewGroup(
				fields...,
			),
		), nil
	}
	// return nil when done. . mabye in the future return some "done error" so we also can handle errors in the wizard
	return nil, nil
}

func (cli *Client) WizComplete() error {
	cli.Tenant = cli.wiz.Tenant
	cli.Uri = cli.wiz.Uri
	err := cli.Warmup()
	if err != nil {
		return fmt.Errorf("failed to warmup vault: %w", err)
	}
	return nil
	// return map[string]any{
	// 	"tenant": cli.wiz.Tenant,
	// 	"uri":    cli.wiz.Uri,
	// }, nil
}
