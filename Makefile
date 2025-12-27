# =========================================================================================== #
# HELPERS
# =========================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


# =========================================================================================== #
# DEVELOPMENT
# =========================================================================================== #

## run/api: run the API application
.PHONY: run/api
run/api:
	go run ./cmd/api

## run/docker/api: run the docker container
.PHONY: run/docker/api
run/docker/api:
	docker run -p 8080:8080 vladgrskkh/mrs-api:latest


# =========================================================================================== #
# QUALITY CONTROL
# =========================================================================================== #
 
## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	golangci-lint run
	@echo 'Running tests...'
	go test -race -vet=off ./...
 
## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# =========================================================================================== #
# BUILD
# =========================================================================================== #
 
current_time := $(shell date -Iseconds)
git_description := $(shell git describe --always --dirty --tags --long)
# linker_flags := '-s -X main.buildTime=$(current_time) -X main.version=$(git_description)'
linker_flags := '-s'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	sed -i '' 's/^API_TODO_VERSION=.*/API_TODO_VERSION=$(git_description)/' .env
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

## build/docker/api: build the docker image and push it to docker hub
.PHONY: build/docker/api
build/docker/api:
	@echo 'Building docker image...'
	sed -i '' s/^API_TODO_VERSION=.*/API_TODO_VERSION=$(git_description)/' .env
	docker build --build-arg LINKER_FLAGS=${linker_flags} --platform linux/amd64,linux/arm64 --tag vladgrskkh/todo .
	@echo 'Pushing docker image...'
	docker push vladgrskkh/todo:latest
