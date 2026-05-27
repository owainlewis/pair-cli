BINARY := pair
PACKAGE := ./cmd/pair
DIST := dist

.PHONY: build test cross-build clean

build:
	go build -o $(BINARY) $(PACKAGE)

test:
	go test ./...

cross-build:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 $(PACKAGE)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64 $(PACKAGE)
	GOOS=linux GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64 $(PACKAGE)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 $(PACKAGE)

clean:
	rm -rf $(BINARY) $(DIST)
