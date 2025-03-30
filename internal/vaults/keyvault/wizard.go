package keyvault

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/charmbracelet/huh"
)

var formGroup int

// func (cli *KeyvaultClient) WizardForm() *huh.Group {

// func
func (cli *KeyvaultClient) WizardNext() *huh.Form {
	// automatically increment the formGroup
	defer func() { formGroup++ }()

	switch formGroup {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[Tenant]().
					Title("Select a tenant").
					OptionsFunc(cli.wiz.tenantOptions, nil).
					Value(&cli.wiz.selectedTenant),
			),
		)
	case 1:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[Subscription]().
					TitleFunc(func() string {
						val := cli.wiz.subs[cli.wiz.selectedTenant]
						return fmt.Sprintf("Select subscriptions (%d)", len(val))
					}, nil).
					Description("Multiple subscriptions can be selected.").
					OptionsFunc(cli.wiz.subscriptionOptions, nil).
					Value(&cli.wiz.selectedSub),
			),
		)
	case 2:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[GraphQueryItem]().
					Title("Select a keyvault").
					OptionsFunc(cli.wiz.keyvaultOptions, nil).
					Value(&cli.wiz.selectedRes),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to include certificates and keys in the .env file?").
					Affirmative("Yes").
					Negative("No").
					Value(&cli.wiz.IncludeCert),
			),
		)
	}
	return nil
}

// deliver selection choices to the wizard
func (wiz *wizard) tenantOptions() []huh.Option[Tenant] {
	slog.Debug("getting tenants")
	wiz.subdone.Wait()
	options := make([]huh.Option[Tenant], 0)
	for t := range wiz.subs {
		opt := huh.NewOption(t.DisplayName, t)
		options = append(options, opt)
	}
	return options
}

func (wiz *wizard) subscriptionOptions() (opts []huh.Option[Subscription]) {
	slog.Debug("getting subscriptions")

	//will be closed when subscriptions are done.
	wiz.subdone.Wait()
	// wiz.resDone.Wait()

	val, ok := wiz.subs[wiz.selectedTenant]
	if !ok {
		slog.Warn("no subscriptions found for tenant", "tenant", wiz.selectedTenant.DisplayName)
		return
	}

	slog.Debug("subs", "tenant", wiz.selectedTenant.DisplayName, "count", len(val))

	opts = make([]huh.Option[Subscription], 0)
	for _, s := range val {
		opt := huh.NewOption(s.DisplayName, *s)
		opts = append(opts, opt)
	}

	return opts
}

func (wiz *wizard) keyvaultOptions() (opts []huh.Option[GraphQueryItem]) {
	slog.Debug("getting keyvaults")
	wiz.resDone.Wait()

	opts = make([]huh.Option[GraphQueryItem], 0)

	for _, res := range wiz.res {
		isInSelectedSub := slices.ContainsFunc(wiz.selectedSub, func(v Subscription) bool {
			return v.Id == res.SubscriptionId
		})
		if !isInSelectedSub {
			continue
		}
		sub := wiz.subs[wiz.selectedTenant]
		i := slices.IndexFunc(sub, func(v *Subscription) bool {
			return v.Id == res.SubscriptionId
		})
		Ssub := sub[i]
		opt := huh.NewOption(fmt.Sprintf("%s (%s)", res.Name, Ssub.DisplayName), *res)
		opts = append(opts, opt)
	}

	return opts
}
