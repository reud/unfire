
docker-test: ./tests/Dockerfile .env
	docker build . -f ./tests/Dockerfile -t unfire-test
	docker run --env-file .env -t unfire-test

start: Dockerfile .env
	docker build . -f ./Dockerfile -t unfire
	docker run --env-file .env -t unfire