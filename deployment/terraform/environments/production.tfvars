# Production Environment Configuration

environment = "production"
aws_region  = "us-west-2"

# Domain configuration
domain_name = "yourdomain.com"
subdomain   = ""

# VPC Configuration
vpc_cidr           = "10.0.0.0/16"
enable_nat_gateway = true

# ECS Configuration
ecs_task_cpu       = 1024
ecs_task_memory    = 2048
ecs_desired_count  = 3
ecs_min_capacity   = 2
ecs_max_capacity   = 10

# RDS Configuration
db_instance_class           = "db.t3.small"
db_allocated_storage        = 50
db_max_allocated_storage    = 200
db_backup_retention_period  = 14
db_multi_az                = true
db_deletion_protection     = true

# ElastiCache Configuration
redis_node_type        = "cache.t3.small"
redis_num_cache_nodes  = 2

# CloudWatch Configuration
log_retention_days = 30

# Auto Scaling Configuration
cpu_target_value    = 60
memory_target_value = 70

# Security Configuration
allowed_cidr_blocks = ["0.0.0.0/0"]  # Restrict this in real production
enable_waf         = true

# Backup Configuration
backup_schedule        = "cron(0 2 * * ? *)"  # Daily at 2 AM UTC
backup_retention_days  = 30