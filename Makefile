.PHONY: all test smoke run lint

all: lint test smoke

test:
	go test -race ./...

smoke:
	go run ./cmd/cachebench -smoke -out results/smoke.json

run:
	go run ./cmd/cachebench -capacities 500,2000 -out results/results.json

lint:
	go vet ./...
