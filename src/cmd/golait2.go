package main

import (
	"flag"

	"github.com/craigmj/commander"
	"github.com/golang/glog"
	`github.com/juju/errors`

	"golait2/parser"
)

func main() {
	flag.Parse()
	if err := commander.Execute(
		flag.Args(),
		parser.GenerateCommand,
	); nil != err {
		glog.Error(errors.ErrorStack(err))
		glog.Fatal(err)
	}
}
