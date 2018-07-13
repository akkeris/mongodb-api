
REGISTRY=$(MYREGISTRY)
ORG_NAME=akkeris
APPNAME=mongodb-api
VERSION=1.0
TAG=$(VERSION)-DEV
REPO=$(REGISTRY)/$(ORG_NAME)/$(APPNAME)
IMAGE=$(REGISTRY)/$(ORG_NAME)/$(APPNAME):$(TAG)
PORT=4040

SRC=*.go
PKGS=mongodb-api/server mongodb-api/db
PKGS_BUILD_ARGS = --build-arg PKGS="$(PKGS)"

DOCKERFILE=Dockerfile

VAULT_BUILD_ARGS = --build-arg VAULT_ADDR=$(VAULT_ADDR) --build-arg VAULT_TOKEN=$(VAULT_TOKEN)
MONGODB_SECRET_ARG = --build-arg MONGODB_SECRET=${MONGODB_SECRET}

BUILD_ARGS = $(VAULT_BUILD_ARGS) $(PKGS_BUILD_ARGS) $(MONGODB_SECRET_ARG)

.PHONY: init test docker

build: $(APPNAME)

$(APPNAME):  Gopkg.lock test $(SRC)
	go build

test: $(SRC)
	go test -v -cover $(PKGS)

Gopkg.lock: Gopkg.toml $(SRC)
	dep ensure

dep:
	dep ensure

init:
	dep init

docker: Dockerfile $(SRC)
	docker build -t $(APPNAME):$(TAG) $(BUILDARGS) $(BUILD_ARGS) -f ${DOCKERFILE} .
	docker tag $(APPNAME):$(TAG) $(APPNAME):dev


all: $(APPNAME) docker

clean: clean_docker clean_app

clean_app:
	-rm -rf $(APPNAME) vendor Gopkg.lock

clean_docker:
	-docker rmi $(APPNAME)
