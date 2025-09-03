package model

type Source interface {
	Get(args []byte, file string) error
}

// convert secret to output bytes
type Formatter interface {
	Format([]StoredEnv) ([]byte, error)
}

// output formatted bytes to a chosen channel
type Emitter interface {
	Emit([]byte) error
}
