# Biclomap Back-End

This is the back-end application, intended to work as an AWS serverless application written in Python

The server-side resources creation is handled using Terraform.

# Setting-up local dev environment

This command will creat the local environment, download the required packages
then start a local uvicorn server instance:
```shell
$ make uvicorn
```

# Deploying to the AWS Environment

You need to set-up first AWS credentials under the profile `biclomap` then
issue this command:
```shell
$ make aws-apply
```

