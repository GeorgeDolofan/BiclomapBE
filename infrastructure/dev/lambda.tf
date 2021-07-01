data "archive_file" "app" {
  type        = "zip"
  output_path = "../../bin/lambda.zip"
  source_file = "../../bin/lambda"
}

resource "aws_lambda_function" "app" {
  function_name = "biclomap_app"
  filename      = "../../bin/lambda.zip"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "lambda"

  source_code_hash = data.archive_file.app.output_base64sha256

  runtime     = "go1.x"
  timeout     = 15
  memory_size = 128

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.example,
  ]
}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "apigateway.amazonaws.com"

  # The "/*/*" portion grants access from any method on any resource
  # within the API Gateway REST API.
  source_arn = "${aws_api_gateway_rest_api.biclomap.execution_arn}/*/*"
}

# This is to optionally manage the CloudWatch Log Group for the Lambda Function.
# If skipping this resource configuration, also add "logs:CreateLogGroup" to the IAM policy below.
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/biclomap_app"
  retention_in_days = 14
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_iam_policy" "lambda_dynamodb_access_facebook_users" {
  name        = "dynamodb_access_facebook_users"
  path        = "/"
  description = "IAM policy to access dynamodb"

  policy = jsonencode({
    Version : "2012-10-17"
    Statement : [{
      Effect : "Allow",
      Action : [
        "dynamodb:BatchGetItem",
        "dynamodb:GetItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:BatchWriteItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem"
      ],
      "Resource" : aws_dynamodb_table.facebook-users.arn
      }
    ]
    }
  )

}

resource "aws_iam_role_policy_attachment" "lambda_db_facebook_users" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.lambda_dynamodb_access_facebook_users.arn
}

resource "aws_iam_policy" "lambda_dynamodb_access_sessions" {
  name        = "dynamodb_access_sessions"
  path        = "/"
  description = "IAM policy to access dynamodb"

  policy = jsonencode({
    Version : "2012-10-17"
    Statement : [{
      Effect : "Allow",
      Action : [
        "dynamodb:BatchGetItem",
        "dynamodb:GetItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:BatchWriteItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem"
      ],
      "Resource" : aws_dynamodb_table.biclomap-sessions.arn
      }
    ]
    }
  )

}

resource "aws_iam_role_policy_attachment" "lambda_db_sessions" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.lambda_dynamodb_access_sessions.arn
}
