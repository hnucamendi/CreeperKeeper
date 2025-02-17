## CreeperKeeper APIGW setup ##
resource "aws_lambda_function" "controller" {
  function_name = var.ck_app_name
  role          = aws_iam_role.main.arn
  architectures = ["x86_64"]
  filename      = "./bootstrap.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
}

# IAM Role
resource "aws_iam_role" "main" {
  name               = "${var.ck_app_name}-role"
  assume_role_policy = data.aws_iam_policy_document.main.json
}

data "aws_iam_policy_document" "main" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com", "scheduler.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

# IAM Role Policies
resource "aws_iam_role_policy" "main" {
  name = "${var.ck_app_name}-role-policy"
  role = aws_iam_role.main.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ],
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "execute-api:ManageConnections",
          "execute-api:Invoke",
        ],
        Resource = "*"
      },
      {
        Action = [
          "ec2:DescribeInstances",
          "ec2:StartInstances",
          "ec2:StopInstances",
        ],
        Effect   = "Allow",
        Resource = "*"
      },
      {
        Effect = "Allow",
        Action = [
          "ssm:GetParameters",
          "ssm:GetParameter",
          "ssm:SendCommand",
        ],
        Resource = [
          "*",
        ]
      },
      {
        Effect = "Allow",
        Action = [
          "dynamodb:PutItem",
          "dynamodb:Scan",
        ],
        Resource = [
          aws_dynamodb_table.main.arn,
        ]
      },
      {
        Effect = "Allow",
        Action = [
          "lambda:InvokeFunction"
        ],
        Resource = [
          aws_lambda_function.controller.arn,
          aws_lambda_function.ec2_monitor.arn,
        ]
      },
    ]
  })
}

# Lambda Permissions for API Gateway
resource "aws_lambda_permission" "main" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.controller.arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "main" {
  name              = "/aws/apigateway/${aws_apigatewayv2_api.main.name}"
  retention_in_days = 7
}

## EC2 register serice setup ##
resource "aws_lambda_function" "ec2_monitor" {
  function_name = "${var.ck_app_name}-ec2-register"
  role          = aws_iam_role.main.arn
  architectures = ["x86_64"]
  filename      = "./bootstrap.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
}

resource "aws_lambda_permission" "ec2_monitor" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.ec2_monitor.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.ec2_monitor.arn
}
