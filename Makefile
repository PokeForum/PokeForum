.PHONY: gen docs

gen:
	go generate ./ent

docs:
	swag init -g cmd/server.go -o docs