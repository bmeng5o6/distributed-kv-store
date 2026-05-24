build:
	docker-compose build

run:
	docker-compose down -v && docker-compose up

fresh:
	docker-compose down -v && docker-compose up --build

test:
	go test ./...

bench:
	go run cmd/benchmark/main.go

clean:
	docker-compose down -v