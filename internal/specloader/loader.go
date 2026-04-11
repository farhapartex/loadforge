package specloader

import "github.com/farhapartex/loadforge/internal/config"

type Loader interface {
	Name() string
	Load(input Input) (*config.Config, error)
}

type Input struct {
	URL      string
	Token    string
	Data     []byte
	Filename string
	Workers  int
	Duration string
	Profile  string
}
