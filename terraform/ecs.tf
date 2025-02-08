// # Local variables
// locals {
// //  ecs_cluster_name        = "CreeperKeeper"
// //  ecs_task_name           = "minecraft-task"
//   minecraft_ami           = "ami-06b21ccaeff8cd686"
//   minecraft_instance_type = "t3.medium"
// }
// 
// # Create ECS Cluster
// // resource "aws_ecs_cluster" "main" {
// //   name = local.ecs_cluster_name
// // 
// //   setting {
// //     name  = "containerInsights"
// //     value = "enabled"
// //   }
// // }
// 
// # ECS Task Definition - Runs EC2 Server
// // resource "aws_ecs_task_definition" "minecraft" {
// //   family                   = local.ecs_task_name
// //   requires_compatibilities = ["EC2"]
// //   network_mode             = "bridge"
// // 
// //   container_definitions = jsonencode([
// //     {
// //       name      = local.ecs_task_name
// //       image     = "itzg/minecraft-server"
// //       cpu       = 512
// //       memory    = 1024
// //       essential = true
// //       portMappings = [
// //         {
// //           containerPort = 25565
// //           hostPort      = 25565
// //           protocol      = "tcp"
// //         }
// //       ]
// //     }
// //   ])
// // 
// //   volume {
// //     name      = "minecraft-data"
// //     host_path = "/ecs/minecraft-data"
// //   }
// // }
// 
// # Key Pair for SSH Access
// resource "aws_key_pair" "minecraft_key" {
//   key_name   = "minecraft-server-key"
//   public_key = file("~/.ssh/id_ed25519.pub")
// }
// 
// # Security Group for Minecraft EC2 Instance
// resource "aws_security_group" "minecraft_sg" {
//   name        = "minecraft-sg"
//   description = "Allow Minecraft and SSH access"
//   vpc_id      = "vpc-0123b9d0536e94660"
// 
//   ingress {
//     description = "Allow Minecraft TCP"
//     from_port   = 25565
//     to_port     = 25565
//     protocol    = "tcp"
//     cidr_blocks = ["0.0.0.0/0"] # Adjust for more security
//   }
// 
//   ingress {
//     description = "Allow SSH"
//     from_port   = 22
//     to_port     = 22
//     protocol    = "tcp"
//     cidr_blocks = ["${var.home_ip}/32"] # Restrict SSH access
//   }
// 
//   egress {
//     from_port   = 0
//     to_port     = 0
//     protocol    = "-1"
//     cidr_blocks = ["0.0.0.0/0"]
//   }
// }
// 
// # IAM Role for EC2 Instance
// resource "aws_iam_role" "minecraft_instance_role" {
//   name = "minecraft-instance-role"
// 
//   assume_role_policy = jsonencode({
//     Version = "2012-10-17",
//     Statement = [
//       {
//         Effect = "Allow",
//         Principal = {
//           Service = "ec2.amazonaws.com"
//         },
//         Action = "sts:AssumeRole"
//       }
//     ]
//   })
// }
// 
// # IAM Instance Profile
// resource "aws_iam_instance_profile" "minecraft_instance_profile" {
//   name = "minecraft-instance-profile"
//   role = aws_iam_role.minecraft_instance_role.name
// }
// 
// # EC2 Instance for Hosting Minecraft
// resource "aws_instance" "mc_ec2" {
//   ami           = local.minecraft_ami
//   instance_type = local.minecraft_instance_type
// 
//   key_name               = aws_key_pair.minecraft_key.key_name
//   vpc_security_group_ids = [aws_security_group.minecraft_sg.id]
//   iam_instance_profile   = aws_iam_instance_profile.minecraft_instance_profile.name
// 
// user_data=<<-EOF
// #!/bin/bash
// # Log script execution
// exec > /var/log/user_data.log 2>&1
// set -x
// 
// # Update system and install required packages
// yum update -y
// yum install -y docker docker-compose
// 
// # Start and enable Docker
// systemctl start docker
// systemctl enable docker
// 
// # Create Minecraft directory
// mkdir -p /minecraft/direwolf
// cd /minecraft/direwolf || exit
// 
// # Debugging: Ensure we are in the correct directory
// pwd
// ls -la
// 
// # Create docker-compose.yml properly
// cat > docker-compose.yml <<EOL
// services:
//   mc:
//     image: itzg/minecraft-server:latest
//     tty: true
//     stdin_open: true
//     ports:
//       - "25565:25565"
//     environment:
//       EULA: "TRUE"
//       TYPE: "FTB"
//       FTB_MODPACK_ID: "126"
//       FTB_MODPACK_VERSION_ID: "100011"
//       MEMORY: "2G"
//       MAX_PLAYERS: "6"
//       MOTD: "RedCraft"
//       TZ: "EST"
//     volumes:
//       - "./data:/data"
// EOL
// 
// # Debugging: Confirm the file was created
// ls -l docker-compose.yml
// cat docker-compose.yml
// 
// # Run the server using docker-compose
// docker-compose up -d
// EOF
// 
// 
//   tags = {
//     Name = "Minecraft-Server"
//   }
// }
// 
