
docker-test: ./tests/Dockerfile
	docker build . -f ./tests/Dockerfile -t unfire-test
	docker run -t unfire-test