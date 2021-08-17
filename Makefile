
AWS_ENV=dev
API_VERSION=1

PYTHON_ENV_DIR=env
LAMBDA_EXECUTABLE=bin/lambda
SWAGGER_JSON=lambda/docs/swagger.json

GO_SRCS:=$(shell find lambda -type f -name '*.go')

clean:
	rm -rf bin

$(LAMBDA_EXECUTABLE): $(GO_SRCS) $(SWAGGER_JSON)
	mkdir -p '$(@D)'
	go mod download
	GOOS=linux go build -o $@ biclomap-be/lambda

$(SWAGGER_JSON): $(GO_SRCS)
	cd $(@D)/.. && swag init

build: $(LAMBDA_EXECUTABLE) $(SWAGGER_JSON)

test: build
	go test biclomap-be/lambda

.ONESHELL:
deploy-plan: $(LAMBDA_EXECUTABLE)
	cd infrastructure/$(AWS_ENV)
	terraform plan

deploy: $(LAMBDA_EXECUTABLE)
	cd infrastructure/$(AWS_ENV)
	terraform apply $(DEPLOY_ARGS)

