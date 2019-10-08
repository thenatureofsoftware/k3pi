Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/TheNatureOfSoftware/k3pi/cmd.Version=$(Version) -X github.com/TheNatureOfSoftware/k3pi/cmd.GitCommit=$(GitCommit)"

.PHONY: all test dist

all: dist

test:
	go test ./...

dist: test
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3pi
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3pi-darwin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3pi-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3pi-arm64
	CGO_ENABLED=0 GOOS=windows go build -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3pi.exe