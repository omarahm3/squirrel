default: build

fmt: 
	gofmt -s -w .

server: fmt
	go run ./cmd/squirreld

server-race:
	go run -race ./cmd/squirreld

build-server:
	go build ./cmd/squirreld

client: fmt
	go run ./cmd/squirrel

client-race:
	go run -race ./cmd/squirrel

build-client:
	go build ./cmd/squirrel

clean:
	go clean -i -r ./main

build:
	for target in `ls ./cmd`; do \
		$(BUILD_ENV_FLAGS) go build ./cmd/$$target; \
	done
