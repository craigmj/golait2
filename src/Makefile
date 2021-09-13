SHELL=/bin/bash
GOSRC=$(shell find . -name "*.go")

../bin/golait2: $(GOSRC)
	go build -o ../bin/golait2 cmd/golait2.go

