terraform {
  required_version = "~> 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.42"
    }
  }
}

provider "aws" {
  # region              = var.region
  region = "us-west-2"
}
