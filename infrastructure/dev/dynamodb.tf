resource "aws_dynamodb_table" "users" {
  name           = "users"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "UserId"

  attribute {
    name = "UserId"
    type = "S"
  }
  attribute {
    name = "email"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  global_secondary_index {
    name               = "EmailIndex"
    hash_key           = "email"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["password", "salt", "token", "Name"]
  }

  tags = {
    Name        = "users"
    Environment = "dev"
  }
}

resource "aws_dynamodb_table" "biclomap-sessions" {
  name           = "biclomap-sessions"
  billing_mode   = "PROVISIONED"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  tags = {
    Name        = "biclomap-sessions"
    Environment = "dev"
  }
}

