.PHONY: build up down clean

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

clean: down
	docker volume prune -f

migrate:
	docker-compose exec web ./api

test-api:
	curl http://localhost:8000/users

test-frontend:
	open http://localhost:3000
