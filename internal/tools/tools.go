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

// check if v is part of argName. if you meant --argument, but wrote -argument, the value coming from cobra would be "rgument"
// example: if argName is "argument", v should not be "rgument"
func CheckDoubleDashS(v string, argName string) error {
	// ex "argument" would be "rgument"
	if v == argName[1:] {
		return fmt.Errorf("argument error: if you are defining %s you have to use double dash. e.g. --%s", argName, argName)
	}
	return nil
}
