// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package keyvault

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/vaulttest"
)

func TestKeyVault(t *testing.T) {
	mock := &mockAzsecretsClient{}
	vaulttest.TestVault(t, &Client{client: mock}, func() model.Vault {
		return &Client{}
	})
}

type mockAzsecretsClient struct {
	// azsecretsClient
	SetSecretCalled bool
}

func (m *mockAzsecretsClient) NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	page := azsecrets.ListSecretPropertiesResponse{
		SecretPropertiesListResult: azsecrets.SecretPropertiesListResult{
			Value: []*azsecrets.SecretProperties{
				{
					ID:          to.Ptr(azsecrets.ID("https://example.vault.azure.net/secrets/secret1")),
					ContentType: to.Ptr("text/plain"),
					Attributes:  &azsecrets.SecretAttributes{Enabled: to.Ptr(true)},
				},
				{
					ID:          to.Ptr(azsecrets.ID("https://example.vault.azure.net/secrets/secret2")),
					ContentType: to.Ptr("text/plain"),
					Attributes:  &azsecrets.SecretAttributes{Enabled: to.Ptr(true)},
				},
			},
		},
	}
	return runtime.NewPager(runtime.PagingHandler[azsecrets.ListSecretPropertiesResponse]{
		More: func(p azsecrets.ListSecretPropertiesResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, p *azsecrets.ListSecretPropertiesResponse) (azsecrets.ListSecretPropertiesResponse, error) {
			return page, nil
		},
	})
}

func (m *mockAzsecretsClient) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{Value: to.Ptr("value1")},
	}, nil
}

func (m *mockAzsecretsClient) SetSecret(ctx context.Context, name string, parameters azsecrets.SetSecretParameters, options *azsecrets.SetSecretOptions) (azsecrets.SetSecretResponse, error) {
	m.SetSecretCalled = true
	return azsecrets.SetSecretResponse{}, nil
}
