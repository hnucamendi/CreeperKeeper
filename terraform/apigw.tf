locals {
  ck_domain_name = var.ck_app_name
  ck_host_name   = "${local.ck_domain_name}.com"


  ck_app_host_name = "app.${local.ck_host_name}"
  ck_web_host_name = "www.${local.ck_host_name}"
  ck_cdn_host_name = "cdn.${local.ck_host_name}"

  ck_jwt_audience = ["${var.ck_app_name}-resource"]
  ck_jwt_issuer   = "https://dev-bxn245l6be2yzhil.us.auth0.com/"
}

# API Gateway
resource "aws_apigatewayv2_api" "main" {
  name          = var.ck_app_name
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["POST", "GET"]
    allow_origins = ["http://localhost:5173", "https://${local.ck_host_name}", "https://${local.ck_web_host_name}"]
    allow_headers = ["authorization", "access-control-allow-origin", "content-type"]
  }
}

resource "aws_apigatewayv2_authorizer" "main" {
  api_id           = aws_apigatewayv2_api.main.id
  name             = "${var.ck_app_name}-authorizer"
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]

  jwt_configuration {
    audience = local.ck_jwt_audience
    issuer   = local.ck_jwt_issuer
  }
}

resource "aws_apigatewayv2_route" "start" {
  api_id               = aws_apigatewayv2_api.main.id
  route_key            = "POST /start"
  target               = "integrations/${aws_apigatewayv2_integration.main.id}"
  authorization_scopes = ["read:all", "write:all"]
  authorizer_id        = aws_apigatewayv2_authorizer.main.id
  authorization_type   = "JWT"
}

resource "aws_apigatewayv2_route" "stop" {
  api_id               = aws_apigatewayv2_api.main.id
  route_key            = "POST /stop"
  target               = "integrations/${aws_apigatewayv2_integration.main.id}"
  authorization_scopes = ["read:all"]
  authorizer_id        = aws_apigatewayv2_authorizer.main.id
  authorization_type   = "JWT"
}

resource "aws_apigatewayv2_route" "add_instance" {
  api_id               = aws_apigatewayv2_api.main.id
  route_key            = "POST /add"
  target               = "integrations/${aws_apigatewayv2_integration.main.id}"
  authorization_scopes = ["write:all"]
  authorizer_id        = aws_apigatewayv2_authorizer.main.id
  authorization_type   = "JWT"
}

resource "aws_apigatewayv2_route" "get_instances" {
  api_id               = aws_apigatewayv2_api.main.id
  route_key            = "GET /instances"
  target               = "integrations/${aws_apigatewayv2_integration.main.id}"
  authorization_scopes = ["read:all"]
  authorizer_id        = aws_apigatewayv2_authorizer.main.id
  authorization_type   = "JWT"
}

resource "aws_apigatewayv2_route" "get_instances" {
  api_id               = aws_apigatewayv2_api.main.id
  route_key            = "GET /test"
  target               = "integrations/${aws_apigatewayv2_integration.main.id}"
}

resource "aws_apigatewayv2_stage" "main" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = var.ck_app_name
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.main.arn
    format = jsonencode({
      requestId : "$context.requestId",
      ip : "$context.identity.sourceIp",
      caller : "$context.identity.caller",
      user : "$context.identity.user",
      requestTime : "$context.requestTime",
      httpMethod : "$context.httpMethod",
      resourcePath : "$context.resourcePath",
      status : "$context.status",
      protocol : "$context.protocol",
      responseLength : "$context.responseLength",
      requestTimeEpoch : "$context.requestTimeEpoch",
      errorMessage : "$context.error.message"
    })
  }

  default_route_settings {
    logging_level            = "INFO"
    data_trace_enabled       = true
    detailed_metrics_enabled = true
    throttling_burst_limit   = 5000
    throttling_rate_limit    = 10000
  }
}

resource "aws_apigatewayv2_domain_name" "main" {
  domain_name = local.ck_app_host_name
  domain_name_configuration {
    certificate_arn = aws_acm_certificate.main.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
  depends_on = [aws_acm_certificate.main]
}

resource "aws_apigatewayv2_api_mapping" "main" {
  api_id      = aws_apigatewayv2_api.main.id
  domain_name = aws_apigatewayv2_domain_name.main.domain_name
  stage       = aws_apigatewayv2_stage.main.name
}

resource "aws_apigatewayv2_deployment" "main" {
  api_id      = aws_apigatewayv2_api.main.id
  description = "${var.ck_app_name} deployment"
  depends_on = [
    aws_apigatewayv2_route.start,
    aws_apigatewayv2_route.stop
  ]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_integration" "main" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  description            = "${var.ck_app_name} server manager"
  payload_format_version = "2.0"
  integration_method     = "POST"
  integration_uri        = aws_lambda_function.controller.invoke_arn
  depends_on             = [aws_lambda_function.controller]
}
