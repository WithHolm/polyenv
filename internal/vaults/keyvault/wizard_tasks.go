package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// return a slice of tenants available to the user
func getTenants() (out []armsubscriptions.TenantIDDescription, err error) {
	// defer close(chn)
	out = make([]armsubscriptions.TenantIDDescription, 0)
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
			out = append(out, *v)
		}
	}
	return out, nil
}

// return slice of subscriptions from selected tenant.
// remember that user needs to have logged in to the tenant using az before sdk returns any subscriptions (even if the tenant is listed in the tenants list)
func getSubscriptions(tenantId string) (out []armsubscriptions.Subscription, err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
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
			if *v.TenantID == tenantId {
				out = append(out, *v)
			}
		}
	}
	slog.Debug("got subscriptions", "count", len(out))
	return out, nil
}

// return a slice of keyvaults from selected subscriptions
func getKeyvaults(subId []string, tenId string) (out []GraphQueryItem, err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID:                   tenId,
		AdditionallyAllowedTenants: []string{"*"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}
	clientFactory, err := armresourcegraph.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %v", err)
	}
	client := clientFactory.NewClient()

	projections := []string{"name", "resourceGroup", "subscriptionId", "tenantId", "location", "vaultUri = properties.vaultUri"}
	subFilter := make([]string, len(subId))
	for i, sub := range subId {
		subFilter[i] = fmt.Sprintf("'%s'", sub)
	}
	subscriptionflt := strings.Join(subFilter, ",")
	query := fmt.Sprintf("resources| where type == 'microsoft.keyvault/vaults' and subscriptionId in (%s)|project %s", subscriptionflt, strings.Join(projections, ","))
	slog.Debug("query to run", "value", query)

	slog.Debug("query to run", "value", query)
	// get first page. this will also tell us if there are more pages
	res, err := client.Resources(ctx, armresourcegraph.QueryRequest{
		Query: to.Ptr(query),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to finish the request: %v", err)
	}

	// what to expect..
	//https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/resourcegraph/armresourcegraph/client_example_test.go
	for {
		//add data to items
		for _, v := range res.Data.([]interface{}) {
			//marshall to json and then unmarshal to struct.. i wish there was a better way
			jData, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json: %s", err)
			}

			//convert to struct
			var queryItem GraphQueryItem
			err = json.Unmarshal(jData, &queryItem)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal json: %s", err)
			}

			out = append(out, queryItem)
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
			return nil, fmt.Errorf("failed to list keyvault resources: %s", err)
		}
	}

	return out, nil
}

// return a slice of keyvault secrets from selected keyvault
func getKeyvaultKeys(vaultUri string, tenId string) (out []*azsecrets.SecretProperties, err error) {
	// ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID:                   tenId,
		AdditionallyAllowedTenants: []string{"*"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	cli, err := azsecrets.NewClient(vaultUri, cred, nil)
	out, err = listSecrets(cli)
	if err != nil {
		return nil, err
	}
	return out, nil
}
