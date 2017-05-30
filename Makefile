GOPATH := ${PWD}/_vendor:${GOPATH}
export GOPATH

build:
	godep go build&&\
	go install github.com/shaynewang/mirc&&\
	go build -o bin/server server/server.go&&\
	go build -o bin/client client/client.go

run_server: build
	./bin/server

run_client: build
	./bin/client

clean:
	rm bin/*
