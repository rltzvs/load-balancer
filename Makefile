up:
	docker-compose up --build -d

down:
	docker-compose down

logs:
	docker-compose logs -f

bench:
	ab -n 1000 -c 100 http://localhost:8080/