all: clean test

build:
	mkdir -p ./tmp/
	go mod tidy; go mod verify
	go build -o iunzip iunzip.go

test: build
	@echo "***** UNIT TESTS NOT YET PROVIDED *****"
	./iunzip -path=test.zip

clean:
	rm -rf ./tmp/
	rm -f ./iunzip

.PHONY: all build test clean