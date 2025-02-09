locals {
  cp_app_name = "creeperkeeper"
}



module "ftb_server" {
  source  = "hnucamendi/minecraft-server-module/aws"
  version = "0.0.7"

  vpc_id                              = var.vpc_id
  app_name                            = local.cp_app_name
  instance_type                       = "t3.medium"
  minecraft_max_players               = 10
  minecraft_motd                      = "RedCraft"
  minecraft_ops_list                  = "Oldjimmy_"
  minecraft_server_type               = "FTBA"
  minecraft_rcon_cmds_last_disconnect = "stop"
  minecraft_memory_G                  = 3
  minecraft_difficulty_level          = 3
  minecraft_world_name                = "RedCraft"
  minecraft_world_seed                = local.cp_app_name
  ftb_modpack_version_id              = 100011
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
