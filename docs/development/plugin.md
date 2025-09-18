# Plugin system

The actual name might change as "plugin" is a bit vague. but its a way of extending polyenvs IO capabilities.

The system is split in three parts:

1. Source Readers -> Can read from a given source (piped data, file, env, etc)
2. Formatters -> Can transform the data from a byte to polyenv data and back
3. Writers -> Can write the data given by formatters to a given destination (stdout, file, etc)

Polyenv is designed to be extensible. You can easily add new output formats and destinations by creating "Plugins".

## Source Readers

This is not yet implemented, but will be in the future.

## Formatters

formatters have 2 main functions:

- Input-Format: takes a byte slice and returns a polyenv secret/value
  - Detect: detects if the formatter can handle the given data
- Output-Format: takes a polyenv "stored-env" (copy of a value from .env file) and returns a byte slice

### Creating a Formatter

1. add a file named `f_myformatter.go` in `internal/plugin/`
1. implement the `Formatter` interface

#### 1. The Formatter Interface

Your formatter must implement this interface from `internal/model/plugins.go`:

```go
type Formatter interface {
    // returns the name of the formatter
	Name() string
    // detects if the formatter can handle the given data
	Detect(data []byte) bool
    // converts byteslice to a polyenv input data (may change in the future)
	InputFormat(data []byte) (*model.InputData, error)
    // converts the given data to a byte slice
	OutputFormat(data []model.StoredEnv) ([]byte, error)
}
```

please note that both detect and inputformat are meant to be used with Source Readers to automatically detect and convert the data. as source readers are not implemeted properly yet, you can just return nil on inputformat and return false.
<!-- * `Name`: returns the name of the formatter
* `Detect`: detects if the formatter can handle the given data
* `InputFormat`: converts byteslice to a polyenv input data (may change in the future)
* `OutputFormat`: converts the given data to a byte slice -->

Honestly, for now you can just return nil on inputformat and return false on detect as both of these are meant to be used with Source Readers to autoamtically detect and convert the data.

you can take a look at any of the f_*.go files in the plugin directory for examples. dotenv is a good example.

#### 2. Create and Register the Formatter

You can add your new formatter directly in the registries in `internal/plugin/a_main.go`
There are different registries for "input" and "output" formatters.
add a key to the map and a factory function to create the formatter.

```go
// if you want to add a new input formatter
var InputFormatters = map[string]func() model.Formatter{
    // ... other formatters
    "myformat": func() model.Formatter { return &MyFormatter{} }, // <-- Add your formatter
}

var OutputFormatters = map[string]func() model.Formatter{
    // ... other formatters
    "myformat": func() model.Formatter { return &MyFormatter{} }, // <-- Add your formatter
}
```

even if you have multiple forms of the same formatter (like json), you can use the same factory function to create different formatters:

```go
var InputFormatters = map[string]func() model.Formatter{
    // ... other formatters
    "json":     func() model.Formatter { return &JSONFormatter{AsArray: false} },
	"jsonArr":  func() model.Formatter { return &JSONFormatter{AsArray: true} },
}
```

NOTE: there is a relationship between formatters and writers if you want writer to specifically write using your given format, see the [doc](#11-acceptedformats)

#### Format tester

Uses golden files to test if the output of a formatter is correct. this is done by running the formatter on a set of test data and comparing the result to a golden file.

for now only "basic test" is defined however more will come in the future.

```go
[]model.StoredEnv{
    {Key: "A", Value: "1", IsSecret: false},
    {Key: "B", Value: "2", IsSecret: true},
},

```

add a file named `basic.{myformat}.golden` in the `testdata` directory.

the file should contain the output of the formatter in the given format.

example from `basic.json.golden`:

``` json
{
  "A": "1",
  "B": "2"
}
```

### Creating a Writer

A `Writer` takes a byte slice from a `Formatter` and writes it to a destination, like stdout

create a new file in `internal/plugin/w_*.go` and implement the `Writer` interface.

#### 1. The Writer Interface

Your writer must implement this interface from `internal/model/plugins.go`:

```go
type Writer interface {
    Name() string
    AcceptedFormats() (accepted []string, deny []string)
    Write([]byte) error
}
```

##### 1.1 AcceptedFormats()

NOTE: this is implemented, but may change in the future.. i realise when i write this that this sounds muddy and i need to find a better implementation. it works for all given cases so far however primitive it is.

Returns 2 string slices of all accepted and denied formats.

for the accepted formats i support `"*"` as wildcard, which means any format is accepted. but you can also add any format as a new entry in front of the `"*"` to make it select that format by default:

`return []string{"json", "*"},[]string{}` -> will accept any format, but select json as default if no format is given.
`return []string{"*"},[]string{}` -> will select any format. this will be a random value (for now)
`return []string{"*"},[]string{"json"}` -> will select any format, but will not accept json.

This is used to to control exactly what formats the writer can write as:

- Github writer can only accept dotenv as github_env or github_output as github actions only support dotenv format.
- Ots writer on the other hand can accept any formatting, but will use 'pick' formatter by default.

#### 2. Create and Register the Writer

register a string tied to a "factory" function in `internal/plugin/a_main.go`

```go
var Writers = map[string]func() model.Writer{
    // ... other writers
    "mywriter": func() model.Writer { return &MyWriter{} }, // <-- Add your writer
}
```

even if you have multiple forms of the same writer (like github), you can use the same factory function to create different writers:

```go
var Writers = map[string]func() model.Writer{
    // ... other writers
    "github-env": func() model.Writer { return &GithubWriter{typ: GithubWriterTypeEnv} }, 
    "github-out": func() model.Writer { return &GithubWriter{typ: GithubWriterTypeDotEnv} },
}