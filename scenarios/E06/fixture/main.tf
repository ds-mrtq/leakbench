variable "region" { type = string }
variable "db_name" { type = string }
variable "db_auth" {
  type      = string
  sensitive = true
}
provider "aws" { region = var.region }
resource "aws_db_instance" "orders" {
  allocated_storage   = 20
  engine              = "postgres"
  instance_class      = "db.t3.micro"
  db_name             = var.db_name
  username            = "app"
  password            = var.db_auth
  publicly_accessible = false
  skip_final_snapshot = true
}
