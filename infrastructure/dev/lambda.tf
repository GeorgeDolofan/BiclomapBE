data "archive_file" "app" {
  type        = "zip"
  output_path = "../../env/lib/python3.8/site-packages/app.zip"
  source_dir  = "../../env/lib/python3.8/site-packages"
}

resource "aws_lambda_function" "app" {
  function_name = "lambda_app"
  filename      = "../../lambda.zip"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "app.main.handler"

  source_code_hash = data.archive_file.app.output_base64sha256

  runtime = "python3.8"
  timeout = 15

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

