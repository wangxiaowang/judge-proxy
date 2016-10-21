all: build run
build:
	go install
run:
	$(GOPATH)/bin/judge-proxy
