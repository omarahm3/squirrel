default: build

fmt: 
	gofmt -s -w .

server: fmt
	go run ./main/squirreld

build-server:
	go build ./main/squirreld

client: fmt
	go run ./main/squirrel

build-client:
	go build ./main/squirrel

clean:
	go clean -i -r ./main

build:
	for target in `ls ./main`; do \
		$(BUILD_ENV_FLAGS) go build ./main/$$target; \
	done
