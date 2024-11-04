package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

// region structs
type Wizard struct {
	Current             int
	Tenant              Tenant
	Tenants             []Tenant
	TenantChannel       chan []Tenant
	Tenantcount         int
	Subscriptions       []Subscription
	SubscriptionChannel chan []Subscription
	Resource            GraphQueryItem
	ResGraphItems       []GraphQueryItem
	ResGraphChannel     chan []GraphQueryItem
	IncludeCertAndKeys  bool
}

var (
	TenantWG       sync.WaitGroup
	SubscriptionWG sync.WaitGroup
	ResGraphWG     sync.WaitGroup
)

type Subscription struct {
	Id          string
	DisplayName string
	TenantId    string
}

type Tenant struct {
	Id          string
	Type        string
	DisplayName string
}

func (t Tenant) Description() string {
	return fmt.Sprintf("(%s) %s", t.Type, t.DisplayName)
}

type GraphQueryItem struct {
	Name           string
	ResourceGroup  string
	SubscriptionId string
	TenantId       string
	VaultUri       string
	Location       string
}

func (g GraphQueryItem) InTenant(t Tenant) bool {
	return g.TenantId == t.Id
}

//endregion structs

// region Helpers

// create a map of all the settings that we need to write to the .env.vaultopts file. this is what is returned by the wizard
func (w *Wizard) GetWizardMap() map[string]string {
	return map[string]string{
		"NAME":                w.Resource.Name,
		"TENANT":              w.Tenant.Id,
		"URI":                 w.Resource.VaultUri,
		"STYLE":               "nocomments", // TODO: make this a setting. supported to be if i want to support comments in env settings.. mabye?
		"ENV_NAME_TAG":        "dotenvKey",
		"INCLUDE_CERTANDKEYS": fmt.Sprintf("%t", w.IncludeCertAndKeys),
	}
}

// region Helpers

// region Tenants
// start job to grab tenants. used by wizard warmup.
func (wiz *Wizard) StartGetTenants() {
	TenantWG.Add(1)
	chn := make(chan []Tenant)
	go func(w *Wizard, chn chan []Tenant) {
		defer TenantWG.Done()
		tenants, err := taskGetTenants()
		if err != nil {
			slog.Error("failed to get tenants: " + err.Error())
			return
		}
		chn <- tenants
	}(wiz, chn)
	wiz.TenantChannel = chn
}

// sets selected tenant
func (w *Wizard) AnswerTenant(tenantId string) error {
	for _, t := range w.Tenants {
		if t.Id == tenantId {
			w.Tenant = t
			return nil
		}
	}
	return fmt.Errorf("failed to find user selected tenant %s in list of tenants", tenantId)
}

// get all available tenants for the identity
func taskGetTenants() (ret []Tenant, err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}
	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %v", err)
	}
	pager := clientFactory.NewTenantsClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tenants: %s", err)
		}
		for _, v := range page.Value {
			ret = append(ret, Tenant{
				Id:          *v.TenantID,
				Type:        *v.TenantType,
				DisplayName: *v.DisplayName,
			})
		}
	}
	return ret, nil
}

// endregion Tenants

// region Subscriptions
// check if tenant has subscriptions
func (wiz *Wizard) TenantHasSub(t Tenant) bool {
	i := slices.IndexFunc(wiz.Subscriptions, func(s Subscription) bool {
		return s.TenantId == t.Id
	})
	return i >= 0
}

func (wiz *Wizard) GetSubName(subId string) string {
	i := slices.IndexFunc(wiz.Subscriptions, func(s Subscription) bool {
		return s.Id == subId
	})
	if i >= 0 {
		return wiz.Subscriptions[i].DisplayName
	}
	return ""
}

// gets subscriptions for all available tenants.
func (w *Wizard) StartGetSubscriptions() {
	w.Subscriptions = make([]Subscription, 0)
	SubscriptionWG.Add(len(w.Tenants))
	chn := make(chan []Subscription, len(w.Tenants))

	for _, t := range w.Tenants {
		go func(w *Wizard, t Tenant, chn chan []Subscription) {
			TenantWG.Wait()
			defer SubscriptionWG.Done()
			subscriptions, err := w.taskGetSubscriptions(t)
			if err != nil {
				slog.Warn("failed to get subscriptions for tenant", "tenant", t.DisplayName)
			}
			chn <- subscriptions
		}(w, t, chn)
	}

	w.SubscriptionChannel = chn
}

// task to get actual subscriptions. used by StartGetSubscriptions as goroutine -> Warmup
func (w *Wizard) taskGetSubscriptions(tenant Tenant) ([]Subscription, error) {
	ret := make([]Subscription, 0)
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		AdditionallyAllowedTenants: []string{"*"},
		TenantID:                   tenant.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %v", err)
	}

	pager := clientFactory.NewClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list subscriptions: %s", err)
		}
		for _, v := range page.Value {

			ret = append(ret, Subscription{
				Id:          *v.SubscriptionID,
				DisplayName: *v.DisplayName,
				TenantId:    *v.TenantID,
			})
		}
	}

	return ret, nil
}

// endregion Subscriptions

// region Keyvaults
func (wiz *Wizard) AnswerKeyvault(answer string) error {
	// find vault in list of vaults, -1 if none is found
	i := slices.IndexFunc(wiz.ResGraphItems, func(g GraphQueryItem) bool {
		return g.Name == answer
	})
	if i != -1 {
		wiz.Resource = wiz.ResGraphItems[i]
		return nil
	}

	return fmt.Errorf("failed to find user selected vault %s in list of vaults", answer)
}

// gets keyvaults for the selected tenant. will block call if StartGetTenants is not called first
func (wiz *Wizard) StartGetKeyvaults() {
	//im not sure how many tenants we have, meaning i dont know how many channels i can make.. lets figure it out
	//only get tenants that i can read subscription from (meaning i have atleast */read permission)
	tenants := make([]Tenant, 0)
	for i := 0; i < len(wiz.Tenants); i++ {
		if !wiz.TenantHasSub(wiz.Tenants[i]) {
			continue
		}
		tenants = append(tenants, wiz.Tenants[i])
	}

	chn := make(chan []GraphQueryItem, len(tenants))
	wiz.Tenantcount = len(tenants)
	for _, t := range wiz.Tenants {
		if !wiz.TenantHasSub(t) {
			continue
		}
		ResGraphWG.Add(1)

		go func(w *Wizard, t Tenant, chn chan []GraphQueryItem) {
			defer ResGraphWG.Done()
			keyvaults, err := w.taskGetKeyvaults(t)
			if err != nil {
				//TODO add logging here
				slog.Debug("failed to get keyvaults", "tenant", t.DisplayName, "error:", err.Error())
				return
			}
			chn <- keyvaults
			// w.ResGraphItems = append(w.ResGraphItems, keyvaults...)
		}(wiz, t, chn)
	}
	wiz.ResGraphChannel = chn
}

// TODO: refactor?
// task to get actual keyvaults. used by StartGetVaults as goroutine -> Warmup
func (wiz *Wizard) taskGetKeyvaults(tenant Tenant) ([]GraphQueryItem, error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: tenant.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}
	clientFactory, err := armresourcegraph.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %v", err)
	}
	client := clientFactory.NewClient()

	//TODO: make this a stuct or something?
	projections := []string{"name", "resourceGroup", "subscriptionId", "tenantId", "location", "vaultUri = properties.vaultUri"}
	query := fmt.Sprintf("resources| where type == 'microsoft.keyvault/vaults'|project %s", strings.Join(projections, ","))
	slog.Debug("query to run", "value", query)

	// get first page. this will also tell us if there are more pages
	res, err := client.Resources(ctx, armresourcegraph.QueryRequest{
		Query: to.Ptr(query),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to finish the request: %v", err)
	}
	// 	// what to expect..
	// 	//https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/resourcegraph/armresourcegraph/client_example_test.go
	ret := make([]GraphQueryItem, 0)
	for {
		//add data to items
		for _, v := range res.Data.([]interface{}) {
			//marshall to json and then unmarshal to struct.. i wish there was a better way
			jData, err := json.Marshal(v)
			if err != nil {
				return ret, fmt.Errorf("failed to marshal json: %s", err)
			}
			var queryItem GraphQueryItem
			err = json.Unmarshal(jData, &queryItem)
			if err != nil {
				return ret, fmt.Errorf("failed to unmarshal json: %s", err)
			}

			ret = append(ret, queryItem)
		}

		if res.SkipToken == nil {
			break
		}
		//get next page
		res, err = client.Resources(ctx, armresourcegraph.QueryRequest{
			Query: to.Ptr(query),
			Options: &armresourcegraph.QueryRequestOptions{
				SkipToken: res.SkipToken,
			},
		}, nil)

		if err != nil {
			return ret, fmt.Errorf("failed to list keyvault resources: %s", err)
		}
	}

	slog.Debug("got keyvaults from tenant", "tenant", tenant.DisplayName, "len", len(ret))
	return ret, nil
}

// endregion Keyvaults
