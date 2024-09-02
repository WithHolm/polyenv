package tools

import "fmt"

// checks if v is part of argName. if you meant --argument, but wrote -argument, the value would be "rgument"
func CheckDoubleDashS(v string, argName string) error {
	if v == argName[1:] {
		return fmt.Errorf("argument error: if you are defining %s you have to use double dash. e.g. --%s", argName, argName)
	}
	return nil
}
