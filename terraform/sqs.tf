resource "aws_sqs_queue" "main" {
  name                      = "${var.ck_app_name}-queue"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dead.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue" "dead" {
  name = "${var.ck_app_name}-dead-queue"
}

resource "aws_sqs_queue_redrive_allow_policy" "main" {
  queue_url = aws_sqs_queue.dead.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.main.arn]
  })
}

resource "aws_lambda_event_source_mapping" "event_source_mapping" {
  event_source_arn = aws_sqs_queue.main.arn
  enabled          = true
  function_name    = aws_lambda_function.sqs_stop.arn
  batch_size       = 1
}
