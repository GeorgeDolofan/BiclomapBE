
AWS_ENV=dev
LAMBDA_ZIP_FILE = lambda.zip

.SECONDEXPANSION:
$(LAMBDA_ZIP_FILE): $$(shell find app -type f -name '*.py')
	7z a $@ ./env/lib/python3.8/site-packages/*
	7z a $@ ./env/lib64/python3.8/site-packages/*
	zip -g $@ -r app

clean:
	rm $(LAMBDA_ZIP_FILE)

.ONESHELL:
aws-plan: $(LAMBDA_ZIP_FILE)
	cd infrastructure/$(AWS_ENV)
	terraform plan

aws-apply: $(LAMBDA_ZIP_FILE)
	cd infrastructure/$(AWS_ENV)
	terraform apply

uvicorn:
	source env/bin/activate
	uvicorn app.main:app --reload

