package local

import (
	"testing"
)

func TestClientString(t *testing.T) {
	c := &Client{}
	expected := "local"
	if c.String() != expected {
		t.Errorf("Expected String() to return '%s', got '%s'", expected, c.String())
	}
}

func TestClientDisplayName(t *testing.T) {
	c := &Client{}
	expected := "Local Cred Store"
	if c.DisplayName() != expected {
		t.Errorf("Expected DisplayName() to return '%s', got '%s'", expected, c.DisplayName())
	}
}

func TestClientMarshalUnmarshal(t *testing.T) {
	originalClient := &Client{
		Service: "my-test-service",
	}

	marshaledMap := originalClient.Marshal()

	newClient := &Client{}
	err := newClient.Unmarshal(marshaledMap)
	if err != nil {
		t.Fatalf("Unmarshal() returned an unexpected error: %v", err)
	}

	if originalClient.Service != newClient.Service {
		t.Errorf("Service not correctly unmarshaled. Expected '%s', got '%s'", originalClient.Service, newClient.Service)
	}
}

// func TestValidateSecretName(t *testing.T) {
// 	c := &Client{Service: "my-service"}

// 	testCases := []struct {
// 		name      string
// 		secretName string
// 		expectErr  bool
// 	}{
// 		{
// 			name:      "valid name",
// 			secretName: "my-secret",
// 			expectErr:  false,
// 		},
// 		{
// 			name:      "empty name",
// 			secretName: "",
// 			expectErr:  true,
// 		},
// 		{
// 			name: "name too long",
// 			secretName: "a_very_long_secret_name_that_is_definitely_going_to_be_longer_than_the_allowed_255_characters_when_combined_with_the_service_name_so_that_we_can_properly_test_the_validation_logic_and_ensure_it_is_working_as_expected_and_not_allowing_long_names_and_now_it_is_actually_long_enough",
// 			expectErr:  true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			_, err := c.ValidateSecretName(tc.secretName)
// 			if (err != nil) != tc.expectErr {
// 				t.Errorf("Expected error: %v, but got: %v", tc.expectErr, err)
// 			}
// 		})
// 	}
// }

func TestClientList(t *testing.T) {
	c := &Client{}
	secrets, err := c.List()

	if err != nil {
		t.Fatalf("List() returned an unexpected error: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("Expected List() to return an empty slice, but got %d items", len(secrets))
	}

	// Ensure it's not nil, but an empty slice
	if secrets == nil {
		t.Errorf("Expected List() to return an empty slice, not nil")
	}
}
