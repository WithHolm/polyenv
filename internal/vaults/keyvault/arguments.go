package keyvault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// gets tenant from microsoft .wellknwon openid config. supports guid and domain for tenant.
func GetTenant(tenant string) (string, error) {
	//
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0/.well-known/openid-configuration", tenant)
	resp, err := http.Get(url)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var v map[string]any
	json.Unmarshal(body, &v)

	if v["token_endpoint"] == nil {
		return "", fmt.Errorf("failed to get correct openid config for tenant %s", tenant)
	}

	//token_endpoint:https://login.microsoftonline.com/{id}/oauth2/v2.0/token
	l, ok := strings.CutPrefix(v["token_endpoint"].(string), "https://login.microsoftonline.com/")
	if !ok {
		return "", fmt.Errorf("failed to parse token endpoint for %s. expected it to start with https://login.microsoftonline.com/, but it didnt.", tenant)
	}
	l, ok = strings.CutSuffix(l, "/oauth2/v2.0/token")
	if !ok {
		return "", fmt.Errorf("failed to parse token endpoint for %s. expected it to end with /oauth2/v2.0/token, but it didnt.", tenant)
	}
	return l, nil
}
