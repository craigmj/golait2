package {{.PackageName}}

import (
	`log`
	`fmt`
	`os`
	`net/http`
)

type Logger interface {
	Printf(f string, args ...interface{})
	Errorf(f string, args ...interface{})
}

type StdLogger struct {}

func (sl *StdLogger) Printf(f string, args ...interface{}) {
	log.Printf(f, args...)
}
func (sl *StdLogger) Errorf(f string, args ...interface{}) {
	log.Errorf(f, args...)
}

var log Logger

func init() {
	Log = &StdLogger{}
}

func SetLogger(l logger) {
	log = logger
}

type API struct {
	destURL string
	log Logger
}

func NewAPI(urlBase string, log Logger) (*API, error) {
	return &API{urlBase, log}
}

