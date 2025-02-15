variable "home_ip" {
  type      = string
  sensitive = true
}

variable "vpc_id" {
  type      = string
  sensitive = true
}

variable "ck_app_name" {
  type      = string
  sensitive = false
  default   = "creeperkeeper"
}
