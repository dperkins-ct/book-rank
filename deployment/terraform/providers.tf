# Additional provider configuration for us-east-1 (required for CloudFront certificates)
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = "BookRank"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}