# Local variables
locals {
  ecs_cluster_name = "CreeperKeeper"
  ecs_task_name    = "minecraft-task"
  minecraft_ami    = "ami-06b21ccaeff8cd686"
  minecraft_instance_type = "t3.medium"
}

# Create ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = local.ecs_cluster_name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition - Runs EC2 Server
resource "aws_ecs_task_definition" "minecraft" {
  family                   = local.ecs_task_name
  requires_compatibilities = ["EC2"]
  network_mode             = "bridge"

  container_definitions = jsonencode([
    {
      name      = local.ecs_task_name
      image     = "itzg/minecraft-server"
      cpu       = 512
      memory    = 1024
      essential = true
      portMappings = [
        {
          containerPort = 25565
          hostPort      = 25565
          protocol      = "tcp"
        }
      ]
    }
  ])

  volume {
    name      = "minecraft-data"
    host_path = "/ecs/minecraft-data"
  }
}

# Key Pair for SSH Access
resource "aws_key_pair" "minecraft_key" {
  key_name   = "minecraft-server-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Security Group for Minecraft EC2 Instance
resource "aws_security_group" "minecraft_sg" {
  name        = "minecraft-sg"
  description = "Allow Minecraft and SSH access"
  vpc_id      = "vpc-0123b9d0536e94660"

  ingress {
    description = "Allow Minecraft TCP"
    from_port   = 25565
    to_port     = 25565
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Adjust for more security
  }

  ingress {
    description = "Allow SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.home_ip}/32"]  # Restrict SSH access
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# IAM Role for EC2 Instance
resource "aws_iam_role" "minecraft_instance_role" {
  name = "minecraft-instance-role"

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

# IAM Instance Profile
resource "aws_iam_instance_profile" "minecraft_instance_profile" {
  name = "minecraft-instance-profile"
  role = aws_iam_role.minecraft_instance_role.name
}

# EC2 Instance for Hosting Minecraft
resource "aws_instance" "mc_ec2" {
  ami           = local.minecraft_ami
  instance_type = local.minecraft_instance_type

  key_name = aws_key_pair.minecraft_key.key_name
  vpc_security_group_ids = [aws_security_group.minecraft_sg.id]
  iam_instance_profile   = aws_iam_instance_profile.minecraft_instance_profile.name

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install docker -y
              systemctl start docker
              systemctl enable docker
              docker run -d -p 25565:25565 --name minecraft itzg/minecraft-server
              EOF

  tags = {
    Name = "Minecraft-Server"
  }
}

# ECS Service to Run the Task (Optional)
resource "aws_ecs_service" "minecraft_service" {
  name            = "minecraft-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.minecraft.arn
  launch_type     = "EC2"

  desired_count = 1
}
