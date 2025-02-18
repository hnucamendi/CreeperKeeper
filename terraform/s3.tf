resource "aws_s3_bucket" "world_data" {
  bucket = "${var.ck_app_name}-world-data"
}
