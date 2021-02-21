
AWS_ENV=dev
API_VERSION=1

PYTHON_ENV_DIR=env
LAMBDA_EXECUTABLE=bin/lambda

GO_SRCS:=$(shell find lambda -type f -name '*.go')

clean:
	rm -rf bin

$(LAMBDA_EXECUTABLE): $(GO_SRCS)
	mkdir -p '$(@D)'
	go mod download
	GOOS=linux go build -o $@ biclomap-be/lambda

build: $(LAMBDA_EXECUTABLE)

test: build
	exit 1

.ONESHELL:
deploy-plan: $(LAMBDA_EXECUTABLE)
	cd infrastructure/$(AWS_ENV)
	terraform plan

deploy: $(LAMBDA_EXECUTABLE)
	cd infrastructure/$(AWS_ENV)
	terraform apply $(DEPLOY_ARGS)

