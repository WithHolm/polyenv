# styleguide

most of it is in https://google.github.io/styleguide/go/guide

## variables

for variables please dont use one letter variables. minimum is three letters.

```go
//Bad
var a int
func (a *a) b() {
}

//Good
var count int
func (cnt *count) increment() {
}
```

## casing

* camelCase for private stuff
* CamelCase for public stuff outside of the package

## logging

this project uses charm log for "output", but slog for the log cmd
keep it simple, only log to info when something is not correct. but debug things you feel are needed.
you can use --debug flag to enable debug if you need more info. this will also enable 

this might create some issues with context and whatnot.. but it is a good start. if you have suggestions, please open an issue.

```go
import "log/slog"

func main() {
	slog.Info("starting")
}
```
