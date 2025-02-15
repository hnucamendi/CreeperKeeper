resource "aws_dynamodb_table" "main" {
  name           = var.ck_app_name
  billing_mode   = "PAY_PER_REQUEST"

  hash_key        = "PK"
  range_key       = "SK"

  attribute {
    name = "PK"
    type = "S"
  }
  attribute {
    name = "SK"
    type = "S"
  }
}
