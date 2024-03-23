export GO_ENV=local

.PHONY: encrypt # env local | prod
encrypt:
	sops --encrypt $(env).secrets.env > $(env).secrets.enc.env

.PHONY: decrypt
decrypt:
	sops --decrypt $(env).secrets.enc.env > $(env).secrets.env

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
	migrate create -ext ".sql" -dir migrations $(name) && rm migrations/**.down.sql

.PHONY: migrate
migrate:
	migrate -source file://migrations -database postgres://postgres:jesse@localhost:5432/quorum?sslmode=disable up

.PHONY: dockerbuild
dockerbuild:
	docker build .



# gcloud auth configure-docker australia-southeast1-docker.pkg.dev