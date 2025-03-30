package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// // returns whatever data as json, indented. for readability and debugging
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
	return pretty.String() // string(pretty.Bytes())
}

// check if string input is part of argName. if you meant --argument, but wrote -argument, the value coming from cobra would be "rgument"
// example: if argName is "argument", v should not be "rgument"..
func CheckDoubleDashS(input string, argName string) error {
	// ex "argument" would be "rgument"
	if input == argName[1:] {
		return fmt.Errorf("argument error: if you are defining %s you have to use double dash. e.g. --%s", argName, argName)
	}
	return nil
}
