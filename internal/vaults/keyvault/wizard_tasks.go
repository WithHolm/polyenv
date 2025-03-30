package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type (
	wizard struct {
		tenantChan     chan *Tenant
		tenantDone     sync.WaitGroup
		selectedTenant Tenant

		subChan     chan *Subscription
		subdone     sync.WaitGroup
		subs        map[Tenant][]*Subscription
		selectedSub []Subscription

		// resChan     chan *GraphQueryItem

		resDone     sync.WaitGroup
		res         []*GraphQueryItem
		selectedRes GraphQueryItem

		IncludeCert bool
	}

	Subscription struct {
		Id          string
		DisplayName string
		TenantId    string
	}

	Tenant struct {
		Id          string
		Type        string
		DisplayName string
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

func (t Tenant) Description() string {
	return fmt.Sprintf("(%s) %s", t.Type, t.DisplayName)
}

// var wiz *wizard

// initialize the wizard

func newWizard() *wizard {
	return &wizard{
		tenantChan:  make(chan *Tenant),
		tenantDone:  sync.WaitGroup{},
		subdone:     sync.WaitGroup{},
		subChan:     make(chan *Subscription),
		subs:        make(map[Tenant][]*Subscription),
		selectedSub: make([]Subscription, 0),
		resDone:     sync.WaitGroup{},
		res:         make([]*GraphQueryItem, 0),
	}
}

// start job to gran tenants and subscriptions. used by wizard warmup.
func (wiz *wizard) Run() {
	//tenant generator
	wiz.tenantDone.Add(1)
	go func() {
		defer close(wiz.tenantChan)
		defer wiz.tenantDone.Done()
		// defer close(wiz.tenantDoneChan)
		err := wiz.getTenants()
		if err != nil {
			slog.Error("failed to get tenants: " + err.Error())
			os.Exit(1)
		}
	}()

	// sub generator
	wiz.subdone.Add(1)
	go func() {
		defer wiz.subdone.Done()
		//start subscription generator. actively reads tenantChan for any new tenants and processes them
		for tenant := range wiz.tenantChan {
			wiz.subdone.Add(1)
			wiz.subs[*tenant] = make([]*Subscription, 0)
			go func() {
				defer wiz.subdone.Done()
				sub, _ := wiz.getSubscriptions(tenant)
				wiz.subs[*tenant] = sub
			}()

		}
	}()

	wiz.resDone.Add(1)
	go func() {
		defer wiz.resDone.Done()
		for {
			if len(wiz.selectedSub) > 0 {
				break
			}
		}
		// defer close(wiz.resChan)
		for _, sub := range wiz.selectedSub {
			wiz.resDone.Add(1)
			go func() {
				defer wiz.resDone.Done()
				err := wiz.getKeyvaults(sub)
				if err != nil {
					slog.Error("failed to get keyvaults: " + err.Error())
					os.Exit(1)
				}
			}()
			// go wiz.getKeyvaults(sub)
		}
	}()
}

//region tenant

// get all available tenants for the identity
func (wiz *wizard) getTenants() (err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create client factory: %v", err)
	}

	pager := clientFactory.NewTenantsClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tenants: %s", err)
		}
		for _, v := range page.Value {
			wiz.tenantChan <- &Tenant{
				Id:          *v.TenantID,
				Type:        *v.TenantType,
				DisplayName: *v.DisplayName,
			}
		}
	}
	return nil
}

// region subscription
func (wiz *wizard) TenantHasSub(t Tenant) bool {
	v := wiz.subs[t]
	if v == nil {
		return false
	}
	if len(v) == 0 {
		return false
	}

	return true
}

// gets subscriptions for a selected tenant
func (wiz *wizard) getSubscriptions(ten *Tenant) (s []*Subscription, err error) {
	ctx := context.Background()

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		AdditionallyAllowedTenants: []string{"*"},
		TenantID:                   ten.Id,
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
			s = append(s, &Subscription{
				Id:          *v.SubscriptionID,
				DisplayName: *v.DisplayName,
				TenantId:    *v.TenantID,
			})
		}
	}

	return s, nil
}

func (wiz *wizard) getKeyvaults(sub Subscription) (err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: sub.TenantId,
	})
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}
	clientFactory, err := armresourcegraph.NewClientFactory(cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create client factory: %v", err)
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
		return fmt.Errorf("failed to finish the request: %v", err)
	}

	// what to expect..
	//https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/resourcegraph/armresourcegraph/client_example_test.go
	for {
		//add data to items
		for _, v := range res.Data.([]interface{}) {
			//marshall to json and then unmarshal to struct.. i wish there was a better way
			jData, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal json: %s", err)
			}

			//convert to struct
			var queryItem GraphQueryItem
			err = json.Unmarshal(jData, &queryItem)
			if err != nil {
				return fmt.Errorf("failed to unmarshal json: %s", err)
			}

			//send to chan
			wiz.res = append(wiz.res, &queryItem)
			// wiz.resChan <- &queryItem
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
			return fmt.Errorf("failed to list keyvault resources: %s", err)
		}
	}

	// slog.Debug("got keyvaults from tenant", "tenant", tenant.DisplayName, "len", len(ret))
	return nil
}
