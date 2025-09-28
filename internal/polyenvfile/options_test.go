// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package polyenvfile

import (
	"reflect"
	"testing"
)

func TestVaultOptions_ConvertString(t *testing.T) {
	opt := VaultOptions{
		HyphenToUnderscore: true,
		UppercaseLocally:   true,
	}

	converted := opt.ConvertString("my-secret")
	if converted != "MY_SECRET" {
		t.Errorf("expected 'my-secret' to be converted to 'MY_SECRET', but got '%s'", converted)
	}
}

func TestVaultOptions_GetVaultOptionHelper(t *testing.T) {
	opt := VaultOptions{}
	val := reflect.ValueOf(opt)

	helpers := opt.GetVaultOptionHelper()

	if val.NumField() != len(helpers) {
		t.Errorf("VaultOptions has %d fields, but GetVaultOptionHelper returns %d helpers", val.NumField(), len(helpers))
	}
}
