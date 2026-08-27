.PHONY: all test smoke run lint build

BIN := bin/cachebench

all: lint test smoke

test:
	go test -race ./...

# Built rather than `go run`: the go command stamps vcs.revision into a
# binary it links, and omits it for `go run`. Every results file this repo
# publishes carries the revision that produced it, so the run has to go
# through a build.
build:
	go build -o $(BIN) ./cmd/cachebench

smoke: build
	./$(BIN) -smoke -out results/smoke.json

run: build
	./$(BIN) -capacities 500,2000 -out results/results.json

lint:
	go vet ./...
