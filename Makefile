postgres:
	docker container run --name postgres12 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d -v "$(CURDIR)/data/postgres12/data:/var/lib/postgresql/data" -p 5432:5432 postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	docker run --rm -v "$(CURDIR):/src" -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test lint
