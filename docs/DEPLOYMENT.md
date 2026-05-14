# BookRank Deployment Guide

This guide covers the complete deployment process for BookRank to AWS using Docker containers, Terraform infrastructure as code, and GitHub Actions CI/CD pipelines.

## Architecture Overview

BookRank is deployed using the following AWS architecture:

- **Compute**: ECS Fargate containers for the Go backend and React frontend
- **Database**: RDS PostgreSQL with automated backups
- **Cache**: ElastiCache Redis for caching and sessions
- **Load Balancing**: Application Load Balancer with SSL termination
- **Static Assets**: S3 + CloudFront CDN
- **DNS**: Route 53 with health checks
- **Monitoring**: CloudWatch logs, metrics, and alarms
- **Security**: WAF, Security Groups, KMS encryption
- **CI/CD**: GitHub Actions for automated testing and deployment

## Prerequisites

### Required Tools

1. **AWS CLI** (v2 recommended)
   ```bash
   aws --version
   aws configure
   ```

2. **Terraform** (>= 1.0)
   ```bash
   terraform --version
   ```

3. **Docker** (>= 20.0)
   ```bash
   docker --version
   ```

4. **Node.js** (>= 18)
   ```bash
   node --version
   npm --version
   ```

5. **Go** (>= 1.21)
   ```bash
   go version
   ```

### AWS Setup

1. **Create AWS Account** and configure billing alerts
2. **Configure IAM User** with appropriate permissions:
   - EC2 Full Access
   - ECS Full Access
   - RDS Full Access
   - ElastiCache Full Access
   - S3 Full Access
   - CloudFront Full Access
   - Route 53 Full Access
   - CloudWatch Full Access
   - IAM permissions for role creation
   - Secrets Manager Full Access
   - KMS Full Access

3. **Create S3 bucket for Terraform state**:
   ```bash
   aws s3 mb s3://bookrank-terraform-state-<unique-suffix>
   aws s3api put-bucket-versioning --bucket bookrank-terraform-state-<unique-suffix> --versioning-configuration Status=Enabled
   ```

4. **Create DynamoDB table for Terraform locks**:
   ```bash
   aws dynamodb create-table \
     --table-name bookrank-terraform-locks \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --billing-mode PAY_PER_REQUEST
   ```

## Environment Setup

### 1. Configure Domain (Production Only)

For production deployment, you need a domain registered in Route 53:

1. **Register domain** or transfer existing domain to Route 53
2. **Update Terraform variables** in `deployment/terraform/environments/production.tfvars`
3. **Set domain_name** to your actual domain

### 2. Update Terraform Backend

Edit `deployment/terraform/main.tf` to use your S3 bucket:

```hcl
backend "s3" {
  bucket         = "your-terraform-state-bucket"
  key            = "infrastructure/terraform.tfstate"
  region         = "us-west-2"
  encrypt        = true
  dynamodb_table = "bookrank-terraform-locks"
}
```

### 3. Configure GitHub Secrets

For GitHub Actions CI/CD, configure the following secrets in your repository:

- `AWS_ACCESS_KEY_ID`: AWS access key for deployment
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for deployment
- `SLACK_WEBHOOK_URL`: (Optional) Slack webhook for deployment notifications

## Deployment Methods

### Method 1: Automated Deployment (Recommended)

#### Staging Deployment

1. **Push to main branch** triggers automatic staging deployment
2. **Monitor deployment** in GitHub Actions
3. **Access application** at the ALB URL shown in the action logs

#### Production Deployment

1. **Manual trigger** via GitHub Actions:
   - Go to Actions tab
   - Select "Deploy to AWS" workflow
   - Click "Run workflow"
   - Select "production" environment

### Method 2: Manual Deployment

#### 1. Initial Infrastructure Setup

```bash
# Clone repository
git clone <repository-url>
cd book-rank

# Deploy to staging
./scripts/deployment/deploy.sh -e staging

# Deploy to production (requires domain setup)
./scripts/deployment/deploy.sh -e production
```

#### 2. Subsequent Deployments

```bash
# Deploy only application (skip Terraform)
./scripts/deployment/deploy.sh -e staging --skip-terraform

# Deploy only infrastructure (skip Docker)
./scripts/deployment/deploy.sh -e staging --skip-docker
```

#### 3. Destroy Infrastructure (Careful!)

```bash
# Destroy staging environment
./scripts/deployment/deploy.sh -e staging --destroy

# Destroy production environment (requires confirmation)
./scripts/deployment/deploy.sh -e production --destroy
```

### Method 3: Step-by-Step Manual Deployment

#### 1. Build and Deploy Infrastructure

```bash
cd deployment/terraform

# Initialize Terraform
terraform init

# Create workspace for environment
terraform workspace new staging  # or production

# Plan deployment
terraform plan -var-file="environments/staging.tfvars" -out=staging.tfplan

# Apply deployment
terraform apply staging.tfplan

# Get outputs
terraform output -json > staging-outputs.json
```

#### 2. Build and Push Docker Image

```bash
# Build frontend
cd frontend
npm ci
npm run build
cd ..

# Build backend Docker image
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-west-2.amazonaws.com

docker build -f Dockerfile.production -t bookrank:latest .
docker tag bookrank:latest <account-id>.dkr.ecr.us-west-2.amazonaws.com/bookrank-staging:latest
docker push <account-id>.dkr.ecr.us-west-2.amazonaws.com/bookrank-staging:latest
```

#### 3. Update ECS Service

```bash
aws ecs update-service \
  --cluster bookrank-staging-cluster \
  --service bookrank-staging-service \
  --force-new-deployment \
  --region us-west-2

# Wait for deployment
aws ecs wait services-stable \
  --cluster bookrank-staging-cluster \
  --services bookrank-staging-service \
  --region us-west-2
```

## Configuration

### Environment Variables

The application uses environment-specific configuration files:

- **Staging**: `.env.staging`
- **Production**: `.env.production`

Key configuration areas:

#### Database Configuration
- Connection pooling optimized for each environment
- SSL required in production
- Automatic failover for production RDS

#### Caching Configuration
- Redis clustering for production
- Different TTL values per environment

#### Security Configuration
- JWT secrets stored in AWS Secrets Manager
- Rate limiting configured per environment
- WAF enabled for production

#### Monitoring Configuration
- CloudWatch logs retention varies by environment
- Different alerting thresholds for staging vs production

## Monitoring and Maintenance

### Health Checks

The application provides multiple health check endpoints:

- `/health` - Comprehensive health check with database and cache status
- `/health/ready` - Readiness probe for load balancer
- `/health/live` - Liveness probe for container orchestration

### CloudWatch Dashboard

Access the monitoring dashboard:
```
https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#dashboards:name=bookrank-<environment>-dashboard
```

### Log Access

View application logs:
```bash
aws logs tail /ecs/bookrank-<environment> --follow --region us-west-2
```

### Database Backups

Automated backups are configured:

- **RDS Automated Backups**: Daily with point-in-time recovery
- **Manual Backup Script**: `scripts/deployment/backup.sh`
- **S3 Backup Storage**: Long-term retention with lifecycle policies

### Alerts

SNS alerts are configured for:
- Application errors
- High CPU/memory usage
- Database connection issues
- Load balancer errors
- Health check failures

## Scaling

### Horizontal Scaling

ECS Auto Scaling is configured based on:
- CPU utilization (target: 70% staging, 60% production)
- Memory utilization (target: 80% staging, 70% production)

### Vertical Scaling

To increase container resources:

1. **Update Terraform variables**:
   ```hcl
   ecs_task_cpu    = 1024  # Increase from 512
   ecs_task_memory = 2048  # Increase from 1024
   ```

2. **Apply changes**:
   ```bash
   terraform apply -var-file="environments/staging.tfvars"
   ```

### Database Scaling

For database scaling:

1. **Vertical**: Change `db_instance_class` in Terraform
2. **Read Replicas**: Automatically created for production
3. **Connection Pooling**: Configured via environment variables

## Security

### Network Security

- **VPC**: Isolated network with public/private subnets
- **Security Groups**: Minimal required access rules
- **NAT Gateway**: Secure outbound internet access for private subnets
- **WAF**: Web Application Firewall for production

### Data Security

- **Encryption at Rest**: RDS, ElastiCache, and S3 use KMS encryption
- **Encryption in Transit**: TLS 1.2+ for all connections
- **Secrets Management**: AWS Secrets Manager for sensitive data
- **Certificate Management**: AWS Certificate Manager for SSL/TLS

### Access Control

- **IAM Roles**: Least privilege access for all services
- **Task Roles**: Container-specific permissions
- **MFA**: Required for administrative access (configured separately)

## Troubleshooting

### Common Issues

#### 1. Deployment Fails

```bash
# Check ECS service events
aws ecs describe-services --cluster <cluster> --services <service> --region us-west-2

# Check container logs
aws logs tail /ecs/bookrank-<environment> --region us-west-2
```

#### 2. Database Connection Issues

```bash
# Check RDS status
aws rds describe-db-instances --db-instance-identifier bookrank-<env>-db --region us-west-2

# Check security groups
aws ec2 describe-security-groups --group-ids <sg-id> --region us-west-2
```

#### 3. Load Balancer Issues

```bash
# Check target group health
aws elbv2 describe-target-health --target-group-arn <arn> --region us-west-2

# Check ALB logs in S3
aws s3 ls s3://bookrank-<env>-alb-logs-<suffix>/alb-logs/
```

### Emergency Procedures

#### 1. Rollback Deployment

```bash
# Get previous task definition
aws ecs describe-services --cluster <cluster> --services <service> --region us-west-2

# Update to previous revision
aws ecs update-service --cluster <cluster> --service <service> --task-definition <previous-arn> --region us-west-2
```

#### 2. Scale Down (Maintenance)

```bash
# Scale to minimum
aws ecs update-service --cluster <cluster> --service <service> --desired-count 0 --region us-west-2
```

#### 3. Database Maintenance

```bash
# Create manual snapshot before maintenance
aws rds create-db-snapshot --db-instance-identifier bookrank-<env>-db --db-snapshot-identifier manual-snapshot-$(date +%Y%m%d) --region us-west-2
```

## Cost Optimization

### Staging Environment

- Uses smaller instance types
- Single-AZ deployment
- Shorter backup retention
- Spot instances where possible

### Production Environment

- Right-sized for actual load
- Multi-AZ for high availability
- Longer backup retention
- Reserved instances for predictable workloads

### Monitoring Costs

- Set up AWS Budgets alerts
- Monitor CloudWatch usage
- Review RDS and ElastiCache utilization
- Optimize log retention periods

## Next Steps

After successful deployment:

1. **Configure monitoring alerts** with appropriate notification channels
2. **Set up automated backups** verification
3. **Configure CI/CD pipelines** for your development workflow
4. **Review and adjust scaling policies** based on actual usage
5. **Implement additional security measures** as needed
6. **Set up disaster recovery procedures**

## Support

For deployment issues:

1. Check CloudWatch logs for errors
2. Review Terraform plan output
3. Verify AWS service quotas
4. Check GitHub Actions logs for CI/CD issues
5. Review this documentation for troubleshooting steps

## Security Considerations

- Regularly update base images and dependencies
- Monitor AWS Security Hub recommendations
- Review IAM permissions periodically
- Keep Terraform and provider versions updated
- Implement proper secrets rotation procedures