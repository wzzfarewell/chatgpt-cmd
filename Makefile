# Set the name of the binary
BINARY_NAME=bin/chatgpt-cmd

# Set the Go build command
GOBUILD=go build

# Set the Go build flags
GOFLAGS=-ldflags="-s -w"

# Set the target operating systems and architectures
TARGETS=windows/amd64 windows/arm linux/amd64 linux/arm darwin/amd64

# Build the binary for all target operating systems and architectures
all:
	for target in $(TARGETS); do \
        os=$${target%/*}; \
        arch=$${target#*/}; \
        if [ $$os = "windows" ]; then \
            GOOS=$$os GOARCH=$$arch $(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME)-$$os-$$arch.exe; \
        else \
            GOOS=$$os GOARCH=$$arch $(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME)-$$os-$$arch; \
        fi \
    done

# Clean up the build artifacts
clean:
	rm -f $(BINARY_NAME)-*