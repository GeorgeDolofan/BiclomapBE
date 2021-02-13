data "archive_file" "lambda-ping" {
  type        = "zip"
  source_file = "lambda/src/ping/main.py"
  output_path = "lambda/packages/ping.zip"
}

resource "aws_lambda_function" "ping" {
  function_name = "lambda_ping"
  filename      = "lambda/packages/ping.zip"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "main.lambda_handler"

  source_code_hash = data.archive_file.lambda-ping.output_base64sha256

  runtime = "python3.8"
  timeout = 15

}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.ping.function_name
  principal     = "apigateway.amazonaws.com"

  # The "/*/*" portion grants access from any method on any resource
  # within the API Gateway REST API.
  source_arn = "${aws_api_gateway_rest_api.biclomap.execution_arn}/*/*"
}

