package tools

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateIsoDate(s string) error {
	s = strings.ToUpper(s)
	match, err := regexp.MatchString(`^P\d+[YMW]$`, s)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("expiration must be in ISO 8601 format, e.g. P1Y. Allowed units are Y (for years), M (for months), W (for weeks)")
	}
	return nil
}
