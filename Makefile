DOCKER_USERNAME ?= misterdelle
APPLICATION_NAME ?= invt-logger-reader
GIT_HASH ?= $(shell git log --format="%h" -n 1)

build:
	go build -o ${APPLICATION_NAME}

build-arm:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o ${APPLICATION_NAME}-arm

build-docker-amd:
	docker build --tag ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH} .

build-docker-arm:
	docker buildx build --platform linux/arm64/v8 --output type=docker -t ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH} .
	docker tag ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH} ${DOCKER_USERNAME}/${APPLICATION_NAME}:latest

push:
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH}
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:latest

build-docker-arm-push:
	make build-docker-arm push
