.PHONY: gen docs

gen:
	go generate ./ent

docs:
	swag init -g cmd/server.go -o docs

# Build for Linux ARM64
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pf-linux-arm64 -ldflags '-s -w -extldflags "-static"'

# Build for Linux AMD64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pf-linux-amd64 -ldflags '-s -w -extldflags "-static"'

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pf-linux-amd64 -ldflags '-s -w -extldflags "-static"'
	upx pf-*

# Upx
upx:
	upx pf-*
