// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package plugin

import (
	"os"
)

type StdOutWriter struct{}

func (e *StdOutWriter) Name() string {
	return "stdout"
}

func (e *StdOutWriter) Write(data []byte) error {
	_, err := os.Stdout.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (e *StdOutWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"stats", "*"}, []string{}
}
