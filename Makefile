.PHONY: encrypt
encrypt:
	sops --encrypt .env > .enc.env

.PHONY: decrypt
decrypt:
	sops --decrypt .enc.env > .env

.PHONY: watch
watch:	
	while true; do \
		fd -e .go | entr -rd go run server.go; \
	done;
