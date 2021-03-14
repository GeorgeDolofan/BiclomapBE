data "archive_file" "app" {
  type        = "zip"
  output_path = "../../bin/lambda.zip"
  source_file = "../../bin/lambda"
}

resource "aws_lambda_function" "app" {
  function_name = "lambda_app"
  filename      = "../../bin/lambda.zip"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "lambda"

  source_code_hash = data.archive_file.app.output_base64sha256

  runtime     = "go1.x"
  timeout     = 15
  memory_size = 128

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

