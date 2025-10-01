// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

type Source interface {
	Name() string
	Detect(args []byte) bool
	Get(args []byte, file string) error
}

// convert secret to output bytes
type Formatter interface {
	Name() string
	//detect if the formatter can handle the given data
	Detect(data []byte) bool
	//convert input data
	InputFormat(data []byte) (*InputData, error)
	//convert output data
	OutputFormat([]StoredEnv) ([]byte, error)
}

type InputData struct {
	IsMap    bool
	IsSlice  bool
	IsString bool
	Value    any
}

// output formatted bytes to a chosen channel
type Writer interface {
	Name() string
	//what formats are supported by this writer. * is all, else a list of formats
	// you can enforce a default by adding a value before the * on accepted (ie "json","*")
	// if you just have "*" the selected formatter will be totally random each time
	AcceptedFormats() (accepted []string, deny []string)
	//write to a writer
	Write([]byte) error
}
