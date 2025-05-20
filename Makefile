PROJECT_NAME=emission
BUILD_VERSION=1.1.0

DOCKER_IMAGE=$(PROJECT_NAME):$(BUILD_VERSION)
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
SQLBOILER_CONFIG=sqlboiler.toml

test:
	go test ./pkg/... -cover
build:
	$(GO_BUILD_ENV) go build -v -o $(PROJECT_NAME)-$(BUILD_VERSION).bin main.go

compose_dev: docker
	cd deploy && BUILD_VERSION=$(BUILD_VERSION) docker-compose up --build --force-recreate -d

docker_prebuild: build
	mkdir -p deploy/conf
	mv $(PROJECT_NAME)-$(BUILD_VERSION).bin deploy/$(PROJECT_NAME).bin; \
	cp -R conf deploy/;

docker_build:
	cd deploy; \
	docker build --rm -t $(DOCKER_IMAGE) .;

docker_postbuild:
	cd deploy; \
	rm -rf $(PROJECT_NAME).bin 2> /dev/null;\
	rm -rf conf 2> /dev/null;\

docker: docker_prebuild docker_build docker_postbuild

mock:
	go generate -x -run="mockgen" ./...


