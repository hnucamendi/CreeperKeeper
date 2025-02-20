resource "aws_s3_bucket" "world_data" {
  bucket = "${var.ck_app_name}-world-data"
}

## cloudfront bucket
resource "aws_s3_bucket" "web" {
  bucket = local.ck_host_name
}

resource "aws_s3_bucket_website_configuration" "web" {
  bucket = aws_s3_bucket.web.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}

resource "aws_s3_bucket_public_access_block" "web" {
  bucket = aws_s3_bucket.web.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "web" {
  bucket = aws_s3_bucket.web.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_policy" "web" {
  bucket = aws_s3_bucket.web.id
  policy = data.aws_iam_policy_document.web.json
}

data "aws_iam_policy_document" "web" {
  statement {
    sid       = ""
    effect    = "Allow"
    resources = ["${aws_s3_bucket.web.arn}/*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = [
      "s3:GetObject",
    ]
  }
}

