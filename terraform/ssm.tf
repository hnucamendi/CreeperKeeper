resource "aws_ssm_parameter" "statemanager_jwt_audience" {
  name  = "/${var.ck_app_name}/jwt/client/audience"
  type  = "SecureString"
  value = "changeme"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "statemanager_jwt_client_secret" {
  name  = "/${var.ck_app_name}/jwt/client/secret"
  type  = "SecureString"
  value = "changeme"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "statemanager_jwt_client_id" {
  name  = "/${var.ck_app_name}/jwt/client/id"
  type  = "SecureString"
  value = "changeme"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "statemanager_jwt_client_url" {
  name  = "/${var.ck_app_name}/jwt/client/url"
  type  = "SecureString"
  value = "changeme"

  lifecycle {
    ignore_changes = [value]
  }
}
