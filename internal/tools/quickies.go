package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// returns whatever data as json, indented. for readability and debugging
func ToIndentedJson(data any) string {
	j, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Errorf("failed to marshal json: %s", err))
	}
	var pretty bytes.Buffer
	err = json.Indent(&pretty, j, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to indent json: %s", err))
		// return "", fmt.Errorf("failed to indent json: %s", err)
	}
	return string(pretty.Bytes())
}
