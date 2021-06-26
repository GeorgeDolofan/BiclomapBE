terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  backend "s3" {
    bucket         = "biclomap-be-terraform-state"
    key            = "terraform/state/key"
    region         = "eu-central-1"
    dynamodb_table = "biclomap-be-terraform-state-locks"
  }
}

provider "aws" {
  region  = "eu-central-1"
  profile = var.aws_profile
}

