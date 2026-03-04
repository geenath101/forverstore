build:
	go build -o build/fs
run: build
	./bin/fs
test:
	go test ./... -v