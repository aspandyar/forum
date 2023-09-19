DOCKER_USERNAME ?= forumContainer
APPLICATION_NAME ?= forum
TLS_DIR ?= tls/
GO_ENV_GOROOT := $(shell go env GOROOT)

start:
	touch st.db
	echo "DB_USER=aspandyar" > .env
	echo "DB_PASSWORD=12345678" >> .env
	(cd $(TLS_DIR) && go run $(GO_ENV_GOROOT)/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost)

build:
	docker build --tag ${APPLICATION_NAME} .

run:
	docker run -d -p 4000:4000 --rm --name ${DOCKER_USERNAME} ${APPLICATION_NAME}

stop:
	docker stop ${DOCKER_USERNAME}
