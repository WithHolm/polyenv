// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package keyvault

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

// return a slice of tenants available to the user
func getTenants() (out []armsubscriptions.TenantIDDescription, err error) {
	// defer close(chn)
	out = make([]armsubscriptions.TenantIDDescription, 0)
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}
	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %w", err)
	}
	pager := clientFactory.NewTenantsClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tenants: %w", err)
		}

		for _, v := range page.Value {
			out = append(out, *v)
		}
	}
	return out, nil
}

// return slice of subscriptions from selected tenant.
// remember that user needs to have logged in to the tenant using az before sdk returns any subscriptions (even if the tenant is listed in the tenants list)
func getSubscriptions(tenantID string) (out []armsubscriptions.Subscription, err error) {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}
	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %w", err)
	}
	pager := clientFactory.NewClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list subscriptions: %w", err)
		}
		for _, v := range page.Value {
			if *v.TenantID == tenantID {
				out = append(out, *v)
			}
		}
	}
	slog.Debug("got subscriptions", "count", len(out))
	return out, nil
}

// map [tenant]graph response
var (
	queryCache = make(map[string][]GraphQueryItem)
	cacheMutex sync.RWMutex
)

// return a slice of keyvaults from selected subscriptions
func getKeyvaults(subID string, tenID string) (out []GraphQueryItem, err error) {
	slog.Debug("getting keyvaults", "subscription", subID, "tenant", tenID)

	cacheMutex.RLock()
	cachedResult, ok := queryCache[tenID]
	cacheMutex.RUnlock()

	if ok {
		slog.Debug("using cached keyvaults", "count", len(cachedResult))
		o := make([]GraphQueryItem, 0)
		for _, v := range cachedResult {
			if v.SubscriptionID == subID {
				o = append(o, v)
			}
		}
		return o, nil
	}

	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID:                   tenID,
		AdditionallyAllowedTenants: []string{"*"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}
	clientFactory, err := armresourcegraph.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client factory: %w", err)
	}
	client := clientFactory.NewClient()

	// var err error
	projections := []string{"name", "resourceGroup", "subscriptionId", "tenantId", "location", "vaultUri = properties.vaultUri"}
	// query := fmt.Sprintf("resources| where type == 'microsoft.keyvault/vaults' and subscriptionId == '%s'|project %s", subId, strings.Join(projections, ","))
	query := fmt.Sprintf("resources| where type == 'microsoft.keyvault/vaults'|project %s", strings.Join(projections, ","))

	// get first page. this will also tell us if there are more pages
	res, err := client.Resources(ctx, armresourcegraph.QueryRequest{
		Query: to.Ptr(query),
		Options: &armresourcegraph.QueryRequestOptions{
			ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to finish the request: %w", err)
	}

	// what to expect..
	//https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/resourcegraph/armresourcegraph/client_example_test.go
	for {
		//add data to items
		items, ok := res.Data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected Resource Graph result format; expected ObjectArray, got %T", res.Data)
		}

		for _, v := range items {
			//marshall to json and then unmarshal to struct.. i wish there was a better way
			jData, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json: %w", err)
			}

			//convert to struct
			var queryItem GraphQueryItem
			err = json.Unmarshal(jData, &queryItem)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal json: %w", err)
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
				ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
				SkipToken:    res.SkipToken,
			},
		}, nil)

		if err != nil {
			return nil, fmt.Errorf("failed to list keyvault resources: %w", err)
		}
	}

	cacheMutex.Lock()
	queryCache[tenID] = out
	cacheMutex.Unlock()

	o := make([]GraphQueryItem, 0)
	for _, v := range out {
		if v.SubscriptionID == subID {
			o = append(o, v)
		}
	}
	slog.Debug("got vaults", "count", len(o))

	return o, nil
}
