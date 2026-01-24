# Configure S3 backup destination

resource "dokploy_destination" "backups" {
  name              = "s3-backups"
  access_key        = var.aws_access_key
  secret_access_key = var.aws_secret_key
  bucket            = "my-dokploy-backups"
  region            = "us-east-1"
  endpoint          = "https://s3.amazonaws.com"
  organization_id   = var.organization_id
}

variable "aws_access_key" {
  description = "AWS access key ID"
  type        = string
}

variable "aws_secret_key" {
  description = "AWS secret access key"
  type        = string
  sensitive   = true
}

variable "organization_id" {
  description = "Dokploy organization ID"
  type        = string
}

output "destination_id" {
  value = dokploy_destination.backups.id
}
