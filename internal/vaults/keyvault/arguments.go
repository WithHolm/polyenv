package keyvault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
)

// gets tenant from microsoft .wellknwon openid config. supports guid and domain for tenant.
func GetTenant(tenant string) (string, error) {
	slog.Debug("getting tenant from microsoft .wellknwon openid config", "tenant", tenant)
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0/.well-known/openid-configuration", tenant)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			slog.Error("failed to close response body", "error", e)
		}

	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var v map[string]any
	err = json.Unmarshal(body, &v)
	if err != nil {
		return "", err
	}

	if v["token_endpoint"] == nil {
		return "", fmt.Errorf("failed to get correct openid config for tenant %s", tenant)
	}

	//token_endpoint:https://login.microsoftonline.com/{id}/oauth2/v2.0/token
	l, ok := strings.CutPrefix(v["token_endpoint"].(string), "https://login.microsoftonline.com/")
	if !ok {
		return "", fmt.Errorf("failed to parse token endpoint for %s. expected it to start with https://login.microsoftonline.com/, but it didnt", tenant)
	}
	l, ok = strings.CutSuffix(l, "/oauth2/v2.0/token")
	if !ok {
		return "", fmt.Errorf("failed to parse token endpoint for %s. expected it to end with /oauth2/v2.0/token, but it didnt", tenant)
	}
	return l, nil
}
