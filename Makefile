DOCKER_USERNAME ?= forumContainer
APPLICATION_NAME ?= forum
TLS_DIR ?= tls/
GO_ENV_GOROOT := $(shell go env GOROOT)
COVERAGE_THRESHOLD ?= 95.0

.PHONY: start build run stop test test-cover test-cover-enforce

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


test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-cover-enforce:
	go test ./... -coverprofile=coverage.out
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/{print $$3}' | tr -d '%'); \
	echo "Total coverage: $$total% (required: $(COVERAGE_THRESHOLD)%)"; \
	awk -v total="$$total" -v threshold="$(COVERAGE_THRESHOLD)" 'BEGIN { exit !(total+0 >= threshold+0) }' || \
	( echo "Coverage gate failed"; exit 1 )

test:
	go test ./...
