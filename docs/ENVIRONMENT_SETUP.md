# Environment Configuration Summary

## Quick Start

### 1. Prerequisites Setup
```bash
# Install required tools
brew install aws-cli terraform docker node go jq

# Configure AWS credentials
aws configure

# Install Go tools
make deps
```

### 2. Development Setup
```bash
# Clone and setup
git clone <repository-url>
cd book-rank
make setup

# Start development
make run
```

### 3. Deployment Setup
```bash
# Configure Terraform backend (one-time setup)
# Edit deployment/terraform/main.tf with your S3 bucket

# Deploy to staging
make deploy-staging

# Deploy to production (requires domain setup)
make deploy-production
```

## Environment Configurations

### Staging Environment
- **Purpose**: Pre-production testing and integration
- **Resources**: Minimal cost configuration
- **Domain**: `staging.yourdomain.com` or ALB DNS
- **Database**: Single-AZ RDS t3.micro
- **Cache**: Single Redis node t3.micro
- **Containers**: 1-3 ECS tasks (256 CPU, 512 MB)
- **Backups**: 3-day retention
- **Monitoring**: Basic CloudWatch alerts

### Production Environment  
- **Purpose**: Live application serving users
- **Resources**: High availability configuration
- **Domain**: `yourdomain.com` with Route 53
- **Database**: Multi-AZ RDS t3.small with read replica
- **Cache**: Multi-node Redis cluster t3.small
- **Containers**: 2-10 ECS tasks (1024 CPU, 2048 MB)
- **Backups**: 14-day retention with S3 lifecycle
- **Monitoring**: Comprehensive alerts and dashboards

## Key Files

### Infrastructure
- `deployment/terraform/` - All Terraform infrastructure code
- `deployment/terraform/environments/` - Environment-specific variables
- `deployment/configs/` - Nginx, Prometheus configurations

### CI/CD
- `.github/workflows/ci.yml` - Continuous integration
- `.github/workflows/deploy.yml` - Automated deployment
- `scripts/deployment/deploy.sh` - Manual deployment script

### Docker
- `Dockerfile.production` - Multi-stage production build
- `docker-compose.production.yml` - Production stack
- `docker-compose.yml` - Development stack

### Configuration
- `.env.staging` - Staging environment variables
- `.env.production` - Production environment variables
- `pkg/config/config.go` - Application configuration

## Security Configuration

### Secrets Management
All sensitive data stored in AWS Secrets Manager:
- Database passwords
- JWT secrets  
- Redis auth tokens
- API keys

### Network Security
- Private subnets for application and database
- Public subnets for load balancers only
- Security groups with minimal required access
- NAT gateways for secure outbound access

### Encryption
- RDS encryption at rest with KMS
- ElastiCache encryption in transit and at rest
- S3 server-side encryption
- ALB with SSL/TLS termination

## Monitoring Setup

### Health Checks
- `/health` - Comprehensive health with database/cache status
- `/health/ready` - Kubernetes-style readiness probe
- `/health/live` - Simple liveness check
- Route 53 health checks for production

### CloudWatch Metrics
- Application metrics (requests, errors, latency)
- Infrastructure metrics (CPU, memory, disk)
- Custom business metrics
- Automated alerting via SNS

### Log Aggregation
- Structured JSON logging
- CloudWatch Logs integration
- Log retention policies per environment
- Centralized error tracking

## Scaling Configuration

### Auto Scaling
- ECS Service auto scaling based on CPU/memory
- Target tracking policies
- Scale-in/scale-out cooldowns
- Minimum/maximum capacity limits

### Database Scaling
- Connection pooling optimization
- Read replica for production
- Performance Insights enabled
- Automated backup scaling

## Cost Optimization

### Staging
- t3.micro instances where possible
- Single-AZ deployments
- Shorter backup retention
- Spot instances for non-critical workloads

### Production
- Right-sized instances based on metrics
- Reserved instances for predictable workloads
- S3 lifecycle policies for backups
- CloudWatch cost monitoring

## Maintenance Procedures

### Regular Tasks
- Monitor CloudWatch dashboards
- Review security group rules
- Update container images
- Rotate secrets (automated)
- Review cost reports

### Backup Procedures
- Automated RDS backups
- Manual snapshot before major changes
- S3 backup verification
- Disaster recovery testing

### Update Procedures
- Test in staging first
- Use blue-green deployments
- Monitor health checks during rollout
- Rollback procedures documented

## Troubleshooting

### Common Issues
1. **ECS tasks failing to start**
   - Check CloudWatch logs
   - Verify secrets access
   - Check security groups

2. **Database connection errors**
   - Verify RDS status
   - Check connection pool settings
   - Review security group rules

3. **High response times**
   - Check ECS CPU/memory metrics
   - Review database performance
   - Verify cache hit rates

### Emergency Contacts
- AWS Support (if applicable)
- On-call rotation
- Escalation procedures

## Development Workflow

### Local Development
```bash
# Start local stack
make setup
make docker-up-dev

# Run application
make run

# Run tests
make test-coverage
```

### Feature Development
```bash
# Create feature branch
git checkout -b feature/new-feature

# Develop and test locally
make test
make lint

# Push and create PR
git push origin feature/new-feature
```

### Release Process
```bash
# Merge to main triggers staging deployment
# Manual promotion to production via GitHub Actions
# Monitor deployment via CloudWatch dashboards
```

## Security Checklist

- [ ] AWS credentials properly configured
- [ ] Secrets stored in AWS Secrets Manager
- [ ] Security groups follow principle of least privilege
- [ ] SSL/TLS enabled for all external connections
- [ ] WAF configured for production
- [ ] Regular security updates applied
- [ ] Access logs enabled and monitored
- [ ] Backup encryption verified

## Performance Checklist

- [ ] Auto scaling policies configured
- [ ] Database connection pooling optimized
- [ ] Cache hit rates monitored
- [ ] Static assets served via CDN
- [ ] Response time alerts configured
- [ ] Load testing completed
- [ ] Performance monitoring dashboard setup

## Compliance Checklist

- [ ] Data encryption at rest and in transit
- [ ] Audit logs enabled
- [ ] Access controls documented
- [ ] Backup procedures tested
- [ ] Incident response plan documented
- [ ] Regular security assessments
- [ ] Compliance reporting automated