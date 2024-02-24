.PHONY: encrypt
encrypt:
	sops --encrypt .env > .enc.env

.PHONY: decrypt
decrypt:
	sops --decrypt .enc.env > .env
	
.PHONY: generate-resolvers
generate-resolvers:
	go run github.com/99designs/gqlgen generate

.PHONY: watch
watch:
	ls **/*.go | entr -c go run server.go
