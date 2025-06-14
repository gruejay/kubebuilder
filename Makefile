
build:
	go build -o bin/kubeguide cmd/main.go

.PHONY: run
run:
	go run ./cmd/main.go
