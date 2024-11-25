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
camelCase for private stuff
CamelCase for public stuff
UPPERCASE_SNAKE_CASE for options in vaultopts

## logging
use slog for logging.
keep it simple, only log to info when something is not correct.
you can use debug if you need more info.
```go
import "log/slog"

func main() {
	slog.Info("starting")
}
