build:
	go build main.go

run:
	go run main.go

migration_up:
	migrate -path migrations/postgres -database "postgres://postgres:12345678@localhost/study_db?sslmode=disable" -verbose up
