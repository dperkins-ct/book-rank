# Staging Environment Configuration

environment = "staging"
aws_region  = "us-west-2"

# Domain configuration
domain_name = "yourdomain.com"
subdomain   = "staging"

# VPC Configuration
vpc_cidr           = "10.1.0.0/16"
enable_nat_gateway = true

# ECS Configuration
ecs_task_cpu       = 256
ecs_task_memory    = 512
ecs_desired_count  = 1
ecs_min_capacity   = 1
ecs_max_capacity   = 3

# RDS Configuration
db_instance_class           = "db.t3.micro"
db_allocated_storage        = 20
db_max_allocated_storage    = 50
db_backup_retention_period  = 3
db_multi_az                = false
db_deletion_protection     = false

# ElastiCache Configuration
redis_node_type        = "cache.t3.micro"
redis_num_cache_nodes  = 1

# CloudWatch Configuration
log_retention_days = 7

# Auto Scaling Configuration
cpu_target_value    = 70
memory_target_value = 80

# Security Configuration
allowed_cidr_blocks = ["0.0.0.0/0"]
enable_waf         = false

# Backup Configuration
backup_schedule        = "cron(0 6 * * ? *)"  # Daily at 6 AM UTC
backup_retention_days  = 7