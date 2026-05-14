# BookRank AWS Infrastructure
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Configure S3 backend for state management
  backend "s3" {
    bucket         = "bookrank-terraform-state"
    key            = "infrastructure/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "bookrank-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "BookRank"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# Local values
locals {
  name_prefix = "bookrank-${var.environment}"

  # Network configuration
  vpc_cidr = var.environment == "production" ? "10.0.0.0/16" : "10.1.0.0/16"
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)

  # Common tags
  common_tags = {
    Project     = "BookRank"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }

  # Database configuration
  db_name     = "bookrank"
  db_username = "bookrank"

  # Domain configuration
  domain_name = var.environment == "production" ? "bookrank.yourdomain.com" : "${var.environment}.bookrank.yourdomain.com"
}