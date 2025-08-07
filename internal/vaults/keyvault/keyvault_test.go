package keyvault

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/withholm/polyenv/internal/model"
)

func TestClient_List(t *testing.T) {
	mock := &mockAzsecretsClient{}
	c := &Client{client: mock}

	secrets, err := c.List()
	if err != nil {
		t.Fatalf("List() returned an error: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, but got %d", len(secrets))
	}
}

func TestClient_Pull(t *testing.T) {
	mock := &mockAzsecretsClient{}
	c := &Client{client: mock}

	secret := model.Secret{RemoteKey: "secret1"}
	content, err := c.Pull(secret)
	if err != nil {
		t.Fatalf("Pull() returned an error: %v", err)
	}

	if content.Value != "value1" {
		t.Errorf("expected value to be 'value1', but got '%s'", content.Value)
	}
}

func TestClient_Push(t *testing.T) {
	mock := &mockAzsecretsClient{}
	c := &Client{client: mock}

	content := model.SecretContent{RemoteKey: "secret1", Value: "new-value"}
	err := c.Push(content)
	if err != nil {
		t.Fatalf("Push() returned an error: %v", err)
	}

	if !mock.SetSecretCalled {
		t.Error("SetSecret was not called")
	}
}

type mockAzsecretsClient struct {
	azsecretsClient
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
