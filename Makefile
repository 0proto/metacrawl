GIT_COMMIT=$$(git log --pretty=format:'%h' -n 1)
GIT_TAG=$$(git name-rev --tags --name-only $(git rev-parse HEAD))
BUILD_DATE=$$(date -u '+%Y-%m-%d %H:%M:%S')

all: build

dev: lint test build-dev

deps:
	dep ensure

build:
	go build \
		-ldflags "-X 'main.version=${GIT_TAG}' -X 'main.commit=${GIT_COMMIT}' -X 'main.buildDate=${BUILD_DATE}'" \
		-o app

build-dev:
	go build \
		-race \
		-o app

debug: build-dev
	./app

test:
	go test \
		-cover \
		./domain/...\
		./services/...\
		./transformers/...\
		.

lint:
	golint \
		-set_exit_status \
		./domain/...\
		./services/...\
		./transformers/...\
		.
