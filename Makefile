DOCKER_USERNAME ?= forumContainer
APPLICATION_NAME ?= forum

build:
	docker build --tag ${APPLICATION_NAME} .

run:
	docker run -d -p 8080:8080 ${APPLICATION_NAME} --rm -name ${DOCKER_USERNAME} ${APPLICATION_NAME}

stop:
	docker stop ${DOCKER_USERNAME}
