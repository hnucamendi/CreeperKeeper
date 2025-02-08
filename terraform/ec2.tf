locals {
  cp_app_name = "creeperkeeper"
}

## EC2 Server ##
resource "aws_instance" "main" {
  ami                    = var.ami
  instance_type          = "t3.medium"
  key_name               = aws_key_pair.main.key_name
  vpc_security_group_ids = [aws_security_group.main.id]
  iam_instance_profile   = aws_iam_instance_profile.main.name
  user_data              = <<-EOF
              #!/bin/bash
              sudo yum update -y
              sudo yum install -y docker
            EOF

  tags = {
    Name = "Minecraft-Server"
  }
}

## Key Pair ##
resource "aws_key_pair" "main" {
  key_name   = "${local.cp_app_name}-key-pair"        # Name for your key pair
  public_key = file("~/.ssh/id_ed25519.pub") # Path to your SSH public key
}

## Security Group ##
resource "aws_security_group" "main" {
  name        = "${local.cp_app_name}-sg"
  description = "Allow inbound Minecraft access"
  vpc_id      = "vpc-0123b9d0536e94660" # Replace with your VPC ID

  ingress {
    description = "Allow Minecraft TCP"
    from_port   = 25565
    to_port     = 25565
    protocol    = "tcp"
    cidr_blocks = ["15.230.221.0/24"] # this should limit access to the east coast 
  }

  ingress {
    description = "Allow SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.home_ip}/32"] # Restrict to your IP address
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

## IAM Config ## 
resource "aws_iam_instance_profile" "main" {
  name = "${local.cp_app_name}-iam-instance-profile"
  role = aws_iam_role.main.name
}

resource "aws_iam_role" "main" {
  name = "${local.cp_app_name}-iam-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "ec2.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

## Outputs ##
output "instance_id" {
  value       = aws_instance.main.id
  description = "The ID of the Minecraft server EC2 instance"
}

output "instance_public_ip" {
  value       = aws_instance.main.public_ip
  description = "Public IP address of the Minecraft server EC2 instance"
}
