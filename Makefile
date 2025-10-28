.PHONY: docs

docs:
	swag init -g cmd/server.go -o docs