module "vanilla" {
  source  = "hnucamendi/minecraft-server-module/aws"
  version = "1.0.6"

  vpc_id             = var.vpc_id
  app_name           = var.ck_app_name
  instance_type      = "t3.small"
  minecraft_ops_list = "Oldjimmy_"
  minecraft_memory_G = 1

  security_group_ingress_rules = {
    "allow-all-mc" = {
      description = "Allow Minecraft TCP"
      from_port   = 25565
      to_port     = 25565
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    },
    "allow-host-ssh" = {
      description = "Allow SSH"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = ["${var.home_ip}/32"] # Restrict to your IP address
    }
  }
}

resource "aws_iam_policy" "s3_policy" {
  name        = "${var.ck_app_name}-s3-policy"
  description = "Policy granting S3 permissions for the Minecraft server instance"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid      = "AllowListBucket",
        Effect   = "Allow",
        Action   = ["s3:ListBucket"],
        Resource = "arn:aws:s3:::creeperkeeper-world-data"
      },
      {
        Sid    = "AllowBucketObjectActions",
        Effect = "Allow",
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ],
        Resource = "arn:aws:s3:::creeperkeeper-world-data/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "s3_policy_attachment" {
  role       = "${var.ck_app_name}-iam-role"
  policy_arn = aws_iam_policy.s3_policy.arn
}

## Direwolf modpack ##
module "ftb_server" {
  source  = "hnucamendi/minecraft-server-module/aws"
  version = "1.0.6"

  vpc_id                              = var.vpc_id
  app_name                            = "Tamochimonoyo"
  instance_type                       = "t3.large"
  minecraft_max_players               = 10
  minecraft_motd                      = "Tamo"
  minecraft_ops_list                  = "Oldjimmy_"
  minecraft_server_type               = "FTBA"
  minecraft_rcon_cmds_last_disconnect = "stop"
  minecraft_memory_G                  = 7
  minecraft_difficulty_level          = 3
  minecraft_world_name                = "RedCraft"
  minecraft_world_seed                = var.ck_app_name
  ftb_modpack_version_id              = 100027
  ftb_modpack_id                      = 126

  security_group_ingress_rules = {
    "allow-all-mc" = {
      description = "Allow Minecraft TCP"
      from_port   = 25565
      to_port     = 25565
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    },
    "allow-host-ssh" = {
      description = "Allow SSH"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = ["${var.home_ip}/32"] # Restrict to your IP address
    }
  }
}
