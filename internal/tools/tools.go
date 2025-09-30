// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package tools contains various tools and helpers for polyenv
package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// // returns whatever data as json, indented. for readability and debugging
func ToIndentedJSON(data any) string {
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

// convert keys of a map to a slice of teh value of the key
func MapKeySlice[Map ~map[K]V, K comparable, V any](m Map) []K {
	keys := make([]K, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// finds string f in map m
func InequalFindInMap[Map ~map[string]V, V any](m Map, f string) (value V, found bool) {
	for k, v := range m {
		if strings.EqualFold(k, f) {
			return v, true
		}
	}

	var zero V
	return zero, false
}

// finds string f in slice s with a case insensitive search
func InequalFindInStrSlice(s []string, f string) (found bool, index int) {
	for i, v := range s {
		if strings.EqualFold(v, f) {
			return true, i
		}
	}
	return false, -1
}
