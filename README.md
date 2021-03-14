# Biclomap Back-End

This is the back-end application, intended to work as an AWS serverless application written in GO

The server-side resources creation is handled using Terraform.

# Setting-up local dev environment

Please install `golang` following the official instructions at
https://golang.org

The build is taken care of a `Makefile`. It containes the required commands to
build the main binary file from the GO sources.

# Deploying to the AWS Environment

You need to set-up first AWS credentials under the profile `biclomap` then
issue this command:
```shell
$ make aws-apply
```

