postgresinit:
	docker run --name sneakers_db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -d postgres:15-alpine

postgres:
	docker exec -it sneakers_db psql

createdb:
	docker exec -it sneakers_db createdb --username=root --owner=root sneaker

dropdb:
	docker exec -it sneakers_db dropdb sneaker

migrateup:
	migrate -path db/migrations -database "postgresql://root:password@localhost:5432/sneaker?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://root:password@localhost:5432/sneaker?sslmode=disable" -verbose down

.PHONY: postgresinit