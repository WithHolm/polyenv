package model

type Source interface {
	Detect(args []byte) bool
	Get(args []byte, file string) error
}

// convert secret to output bytes
type Formatter interface {
	//detect if the formatter can handle the given data
	Detect(data []byte) bool
	//convert input data
	InputFormat(data []byte) (any, InputFormatType)
	//convert output data
	OutputFormat([]StoredEnv) ([]byte, error)
}

type InputFormatType int

const (
	InputFormatMap InputFormatType = iota
	InputFormatString
	InputFormatStrSlice
	InputFormatError
)

// output formatted bytes to a chosen channel
type Writer interface {
	//what formats are supported by this writer. * is all, else a list of formats
	AcceptedFormats() (accepted []string, deny []string)
	//write to a writer
	Write([]byte) error
}
