DOCKER_USERNAME ?= forumContainer
APPLICATION_NAME ?= forum

build:
	touch st.db
	docker build --tag ${APPLICATION_NAME} .

run:
	docker run -d -p 4000:4000 --rm --name ${DOCKER_USERNAME} ${APPLICATION_NAME}

stop:
	docker stop ${DOCKER_USERNAME}
