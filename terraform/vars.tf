variable "home_ip" {
  type      = string
  sensitive = true
}

variable "vpc_id" {
  type      = string
  sensitive = true
}

variable "ami" {
  type      = string
  sensitive = false
  default   = "ami-06b21ccaeff8cd686"
}
