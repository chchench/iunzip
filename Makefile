all: clean test

build:
	mkdir -p ./tmp/
	go mod tidy; go mod verify
	go build -o iunzip iunzip.go

test: build
	./iunzip -path=test.zip
	@echo "***** UNIT TESTS NOT YET PROVIDED *****"

clean:
	rm -rf ./tmp/
	rm -f ./iunzip

.PHONY: all build test clean