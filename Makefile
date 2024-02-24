.PHONY: encrypt
encrypt:
	sops --encrypt .env > .enc.env

.PHONY: decrypt
decrypt:
	sops --decrypt .enc.env > .env
	