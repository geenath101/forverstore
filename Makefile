build:
	rm -r build/ || true
	go build -o build/fs
run: 
	rm -r build/ || true
	go build -o build/fs
	./build/fs
test:
	go test ./... -v