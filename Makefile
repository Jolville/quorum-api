export GO_ENV=local

.PHONY: encrypt
encrypt:
	sops --encrypt local.secrets.env > local.secrets.enc.env

.PHONY: decrypt
decrypt:
	sops --decrypt local.secrets.enc.env > local.secrets.env

.PHONY: watch
watch:
	while sleep 1; do \
		fd -e .go | entr -rd go run server.go; \
	done;

.PHONY: dbup
dbup:
	docker pull postgres:15 && docker run --rm --name db_local -d -e POSTGRES_DB=quorum -e POSTGRES_PASSWORD=jesse -p 5432:5432 postgres:15

.PHONY: dbdown
dbdown:
	docker kill db_local

.PHONY: migration
migration:
	migrate create -ext ".sql" -dir migrations $(name)

.PHONY: migrate # direction "up"|"down"
migrate:
	migrate -source file://migrations -database postgres://postgres:jesse@localhost:5432/quorum?sslmode=disable $(direction)
