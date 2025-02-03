build:
	go build -o main cmd/main.go

run:
	./main

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
