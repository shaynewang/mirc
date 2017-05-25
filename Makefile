all:
	go install github.com/shaynewang/mirc&&\
	go build cmd/client/client.go && go build cmd/server/server.go
