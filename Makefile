build:
	go build -o build/fs
run: build
	./build/fs
test:
	go test ./... -v