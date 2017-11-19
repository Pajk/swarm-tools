
VERSION := 0.1.0
SHELL   := /bin/bash

# DOCKER

build:
	docker build -t pajk/swarm-tools:${VERSION} .

publish:
	docker push pajk/swarm-tools:${VERSION}

publish-latest:
	docker tag pajk/swarm-tools:${VERSION} pajk/swarm-tools:latest && docker push pajk/swarm-tools:latest

run:
	docker run --name swarm-tools --rm -p 8080:80 -v /var/run/docker.sock:/var/run/docker.sock pajk/swarm-tools:${VERSION}

# LOCAL DEVELOPMENT

PID      = /tmp/swarm-keeper.pid
GO_FILES = $(wildcard *.go)
APP      = ./app

serve: restart
	@fswatch -o . | xargs -n1 -I{}  make restart || make kill

kill:
	@kill `cat $(PID)` || true

$(APP): $(GO_FILES)
	@go build -o $@ $?

restart: kill $(APP)
	PORT=8000 AUTH_KEY=abcde ./app & echo $$! > $(PID)

.PHONY: build publish run call serve restart kill before