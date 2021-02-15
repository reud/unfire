
docker-test: ./tests/Dockerfile .env
	docker build . -f ./tests/Dockerfile -t unfire-test
	docker run --env-file .env -t unfire-test