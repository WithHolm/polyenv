package keyvault

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
)

type (
	wizSelectType int
	Wizard        struct {
		Tenant       string
		Subscription string
		Uri          string
		Name         string
		state        int
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

const (
	selectTypeKeys wizSelectType = iota
	selectTypeContent
	selectTypeNone
)

// var formGroup int

func (wiz *Wizard) Next() *huh.Form {
	// automatically increment the formGroup
	defer func() { wiz.state++ }()
	// slog.Info("next", "formGroup", formGroup)
	switch wiz.state {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a tenant").
					OptionsFunc(func() []huh.Option[string] {
						v, e := getTenants()
						if e != nil {
							slog.Error("failed to get tenants: " + e.Error())
							os.Exit(1)
						}
						ret := make([]huh.Option[string], 0)
						for _, tenant := range v {
							opt := huh.NewOption(*tenant.DisplayName, *tenant.TenantID)
							ret = append(ret, opt)
						}
						return ret
					}, nil).
					Value(&wiz.Tenant)),
		)
	case 1:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Subscription").
					OptionsFunc(func() (ret []huh.Option[string]) {
						subs, err := getSubscriptions(wiz.Tenant)
						if err != nil {
							slog.Error("failed to get subscriptions: " + err.Error())
							os.Exit(1)
						} else if len(subs) == 0 {
							slog.Error("no subscriptions found. please auth to tenant: az login --tenant " + wiz.Tenant)
							os.Exit(1)
						}
						for _, sub := range subs {
							opt := huh.NewOption(*sub.DisplayName, *sub.SubscriptionID)
							ret = append(ret, opt)
						}
						return ret
					}, wiz.Tenant).Value(&wiz.Subscription),
				huh.NewSelect[string]().
					Title("Select a keyvault").
					OptionsFunc(func() []huh.Option[string] {
						// chn := make(chan armsubscriptions.TenantIDDescription)
						resources, e := getKeyvaults(wiz.Subscription, wiz.Tenant)
						if e != nil {
							slog.Error("failed to get keyvault resources: " + e.Error())
							os.Exit(1)
						}
						ret := make([]huh.Option[string], 0)
						for _, res := range resources {
							opt := huh.NewOption(res.Name, res.VaultUri)
							ret = append(ret, opt)
						}
						return ret
					}, &wiz.Subscription).
					Value(&wiz.Uri),
			),
		)
	}
	return nil
}
