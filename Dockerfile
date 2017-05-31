# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/shaynewang/mirc

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/tools/godep
RUN go install github.com/shaynewang/mirc
RUN cd /go/src/github.com/shaynewang/mirc && make

RUN ["chmod", "+x", "/go/src/github.com/shaynewang/mirc/bin/server"]
# Run the outyet command by default when the container starts.
ENTRYPOINT /go/src/github.com/shaynewang/mirc/bin/server

# Document that the service listens on port 6667
EXPOSE 6667
