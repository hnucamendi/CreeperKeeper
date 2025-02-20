resource "aws_cloudfront_distribution" "app" {
  origin {
    domain_name = "${local.ck_host_name}.s3-website-us-east-1.amazonaws.com"
    origin_id   = var.ck_app_name

    custom_origin_config {
      http_port              = "80"
      https_port             = "443"
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1", "TLSv1.1", "TLSv1.2"]
    }
  }
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "creeperkeeper.com cloudfront distro"
  default_root_object = "index.html"

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }
  aliases = [local.ck_host_name, local.ck_web_host_name]

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = var.ck_app_name

    forwarded_values {
      query_string = true
      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  price_class = "PriceClass_100"

  restrictions {
    geo_restriction {
      locations        = []
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = false
    acm_certificate_arn            = aws_acm_certificate_validation.validation.certificate_arn
    ssl_support_method             = "sni-only"
  }
}

resource "aws_route53_record" "records" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name = dvo.domain_name
      alias = {
        name    = aws_cloudfront_distribution.app.domain_name
        zone_id = aws_cloudfront_distribution.app.hosted_zone_id
      }
    }
  }

  allow_overwrite = true
  name            = each.value.name
  type            = "A"
  zone_id         = data.aws_route53_zone.zone.zone_id

  alias {
    name                   = each.value.alias.name
    zone_id                = each.value.alias.zone_id
    evaluate_target_health = false
  }
}
