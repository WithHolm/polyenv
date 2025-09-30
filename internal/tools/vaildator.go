// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
