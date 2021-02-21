
AWS_ENV=dev
API_VERSION=1
LAMBDA_ZIP_FILE = lambda.zip

APP_ENV_FILE=.env
PYTHON_ENV_DIR=env

.ONESHELL:
.PHONY: $(APP_ENV_FILE)
$(APP_ENV_FILE):
	@echo "Updating .env"
	truncate -s 0 $@
	echo 'VERSION="'$(API_VERSION)'"' > $@
	echo 'RUNTIME_ENVIRONMENT="'$(AWS_ENV)'"' >> $@

$(PYTHON_ENV_DIR):
	virtualenv -p python3.8 env
	source env/bin/activate
	pipenv install $(if $(findstring dev,$(AWS_ENV)),--dev)

.SECONDEXPANSION:
$(LAMBDA_ZIP_FILE): $(PYTHON_ENV_DIR) $(APP_ENV_FILE) $$(shell find app -type f -name '*.py')
	7z a $@ ./env/lib/python3.8/site-packages/*
	7z a $@ ./env/lib64/python3.8/site-packages/*
	zip -g $@ -r app
	zip -g $@ $(APP_ENV_FILE)

clean:
	rm -rf env
	rm -f $(LAMBDA_ZIP_FILE)

aws-plan: $(LAMBDA_ZIP_FILE)
	cd infrastructure/$(AWS_ENV)
	terraform plan

aws-apply: $(LAMBDA_ZIP_FILE)
	cd infrastructure/$(AWS_ENV)
	terraform apply

# this will launch a local webserver. The base URL will be displayed in the
# console output by uvicorn
# In the broweser, use the /docs resource to inspect the API
#
uvicorn: $(APP_ENV_FILE) $(PYTHON_ENV_DIR)
	source env/bin/activate
	uvicorn app.main:app --reload

