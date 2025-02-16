locals {
  ec2_running_monitor_name = "ec2-monitor"
}
resource "aws_scheduler_schedule_group" "main" {
  name = var.ck_app_name
}

resource "aws_cloudwatch_event_rule" "ec2_monitor" {
  name        = "${local.ec2_running_monitor_name}-rule"
  description = "Trigger Lambda when an EC2 instance goes into running state"
  event_pattern = jsonencode({
    "source" : ["aws.ec2"]
    "detail-type" : ["EC2 Instance State-change Notification"]
    "detail" : {
      "state" : ["running", "stopping"]
      "instance-id" : [module.vanilla.instance_id]
    }
  })
}

resource "aws_cloudwatch_event_target" "ec2_monitor" {
  rule      = aws_cloudwatch_event_rule.ec2_monitor.name
  target_id = aws_lambda_function.ec2_monitor.function_name
  arn       = aws_lambda_function.ec2_monitor.arn
}
