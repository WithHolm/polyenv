package output

import "github.com/withholm/polyenv/internal/model"

var Sources = map[string]func() model.Source{
	"ots": func() model.Source { return &OneTimeSecretSource{} },
	"env": func() model.Source { return &EnvSource{} },
}

var Formatters = map[string]func() model.Formatter{
	"json":    func() model.Formatter { return &JsonFormatter{} },
	"jsonArr": func() model.Formatter { return &JsonFormatter{AsArray: true} },
	"pwsh":    func() model.Formatter { return &PwshFormatter{} },
}

var Emitters = map[string]func() model.Emitter{
	"stdout": func() model.Emitter { return &StdoutEmitter{} },
	"ots":    func() model.Emitter { return &OneTimeSecretEmitter{} },
}
