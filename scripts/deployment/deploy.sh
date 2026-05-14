#!/bin/bash
# Deployment Script for BookRank

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/deployment/terraform"

# Default values
ENVIRONMENT=""
AWS_REGION="us-west-2"
SKIP_TERRAFORM=false
SKIP_DOCKER=false
DESTROY=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARN: $1${NC}" >&2
}

error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" >&2
}

success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1${NC}" >&2
}

# Help function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deploy BookRank to AWS

OPTIONS:
    -e, --environment ENV    Environment to deploy (staging|production)
    -r, --region REGION      AWS region (default: us-west-2)
    -s, --skip-terraform     Skip Terraform deployment
    -d, --skip-docker        Skip Docker build and push
    --destroy               Destroy infrastructure (use with caution)
    -h, --help              Show this help message

EXAMPLES:
    $0 -e staging
    $0 -e production -r us-east-1
    $0 -e staging --skip-terraform
    $0 -e staging --destroy

REQUIREMENTS:
    - AWS CLI configured
    - Terraform installed
    - Docker installed
    - jq installed

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -s|--skip-terraform)
            SKIP_TERRAFORM=true
            shift
            ;;
        -d|--skip-docker)
            SKIP_DOCKER=true
            shift
            ;;
        --destroy)
            DESTROY=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$ENVIRONMENT" ]]; then
    error "Environment is required. Use -e staging or -e production"
    usage
    exit 1
fi

if [[ "$ENVIRONMENT" != "staging" ]] && [[ "$ENVIRONMENT" != "production" ]]; then
    error "Environment must be 'staging' or 'production'"
    exit 1
fi

# Check required tools
check_requirements() {
    log "Checking requirements..."

    local missing_tools=()

    if ! command -v aws &> /dev/null; then
        missing_tools+=("aws")
    fi

    if ! command -v terraform &> /dev/null; then
        missing_tools+=("terraform")
    fi

    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi

    if ! command -v jq &> /dev/null; then
        missing_tools+=("jq")
    fi

    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi

    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        error "AWS credentials not configured or invalid"
        exit 1
    fi

    success "All requirements satisfied"
}

# Build and push Docker image
build_and_push_docker() {
    log "Building and pushing Docker image for $ENVIRONMENT..."

    cd "$PROJECT_ROOT"

    # Get AWS account ID and ECR registry
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    ECR_REGISTRY="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"
    ECR_REPOSITORY="bookrank-$ENVIRONMENT"

    # Login to ECR
    log "Logging into ECR..."
    aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_REGISTRY"

    # Create repository if it doesn't exist
    if ! aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" &> /dev/null; then
        log "Creating ECR repository: $ECR_REPOSITORY"
        aws ecr create-repository --repository-name "$ECR_REPOSITORY" --region "$AWS_REGION"
    fi

    # Build frontend first
    log "Building frontend..."
    cd "$PROJECT_ROOT/frontend"
    npm ci
    npm run build

    # Build Docker image
    cd "$PROJECT_ROOT"
    IMAGE_TAG=$(git rev-parse --short HEAD)
    FULL_IMAGE_NAME="$ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG"
    LATEST_IMAGE_NAME="$ECR_REGISTRY/$ECR_REPOSITORY:$ENVIRONMENT-latest"

    log "Building Docker image: $FULL_IMAGE_NAME"
    docker build -f Dockerfile.production -t "$FULL_IMAGE_NAME" .
    docker tag "$FULL_IMAGE_NAME" "$LATEST_IMAGE_NAME"

    # Push Docker image
    log "Pushing Docker image to ECR..."
    docker push "$FULL_IMAGE_NAME"
    docker push "$LATEST_IMAGE_NAME"

    success "Docker image built and pushed successfully"

    # Export image name for Terraform
    export TF_VAR_app_image="$FULL_IMAGE_NAME"
}

# Deploy infrastructure with Terraform
deploy_infrastructure() {
    log "Deploying infrastructure for $ENVIRONMENT with Terraform..."

    cd "$TERRAFORM_DIR"

    # Initialize Terraform
    log "Initializing Terraform..."
    terraform init -upgrade

    # Select workspace
    if ! terraform workspace select "$ENVIRONMENT" 2>/dev/null; then
        log "Creating new workspace: $ENVIRONMENT"
        terraform workspace new "$ENVIRONMENT"
    fi

    # Plan deployment
    log "Planning Terraform deployment..."
    terraform plan \
        -var-file="environments/$ENVIRONMENT.tfvars" \
        -var="aws_region=$AWS_REGION" \
        -out="$ENVIRONMENT.tfplan"

    # Apply deployment
    if [[ "$DESTROY" == "true" ]]; then
        warn "DESTROYING infrastructure for $ENVIRONMENT"
        read -p "Are you sure you want to destroy the infrastructure? Type 'yes' to confirm: " confirm
        if [[ "$confirm" != "yes" ]]; then
            log "Destruction cancelled"
            exit 0
        fi
        terraform destroy \
            -var-file="environments/$ENVIRONMENT.tfvars" \
            -var="aws_region=$AWS_REGION" \
            -auto-approve
        success "Infrastructure destroyed"
    else
        log "Applying Terraform deployment..."
        terraform apply "$ENVIRONMENT.tfplan"

        # Get outputs
        log "Getting Terraform outputs..."
        terraform output -json > "$ENVIRONMENT-outputs.json"

        success "Infrastructure deployed successfully"
    fi
}

# Update ECS service
update_ecs_service() {
    log "Updating ECS service..."

    # Get ECS cluster and service names from Terraform outputs
    CLUSTER_NAME=$(jq -r '.ecs_cluster_name.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")
    SERVICE_NAME=$(jq -r '.ecs_service_name.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")

    if [[ "$CLUSTER_NAME" == "null" ]] || [[ "$SERVICE_NAME" == "null" ]]; then
        error "Could not get ECS cluster or service name from Terraform outputs"
        exit 1
    fi

    # Force new deployment
    log "Forcing new deployment for ECS service: $SERVICE_NAME"
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$SERVICE_NAME" \
        --force-new-deployment \
        --region "$AWS_REGION" > /dev/null

    # Wait for deployment to complete
    log "Waiting for ECS service to stabilize..."
    aws ecs wait services-stable \
        --cluster "$CLUSTER_NAME" \
        --services "$SERVICE_NAME" \
        --region "$AWS_REGION"

    success "ECS service updated successfully"
}

# Run smoke tests
run_smoke_tests() {
    log "Running smoke tests..."

    # Get application URL from Terraform outputs
    APP_URL=$(jq -r '.application_url.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")

    if [[ "$APP_URL" == "null" ]]; then
        warn "Could not get application URL from Terraform outputs, skipping smoke tests"
        return
    fi

    # Wait for application to be ready
    log "Waiting for application to be ready at $APP_URL"
    for i in {1..30}; do
        if curl -f -s "$APP_URL/health" > /dev/null; then
            success "Application is ready"
            break
        fi
        if [[ $i -eq 30 ]]; then
            error "Application failed to become ready after 5 minutes"
            exit 1
        fi
        sleep 10
    done

    # Run basic tests
    log "Testing health endpoint..."
    HEALTH_RESPONSE=$(curl -s "$APP_URL/health")
    if echo "$HEALTH_RESPONSE" | jq -e '.status == "healthy"' > /dev/null; then
        success "Health check passed"
    else
        error "Health check failed: $HEALTH_RESPONSE"
        exit 1
    fi

    success "Smoke tests passed"
}

# Main deployment function
main() {
    log "Starting deployment for environment: $ENVIRONMENT"
    log "AWS Region: $AWS_REGION"

    check_requirements

    if [[ "$DESTROY" == "true" ]]; then
        if [[ "$SKIP_TERRAFORM" == "false" ]]; then
            deploy_infrastructure
        fi
        exit 0
    fi

    if [[ "$SKIP_DOCKER" == "false" ]]; then
        build_and_push_docker
    fi

    if [[ "$SKIP_TERRAFORM" == "false" ]]; then
        deploy_infrastructure
        update_ecs_service
        run_smoke_tests
    fi

    success "Deployment completed successfully!"

    # Display useful information
    if [[ -f "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json" ]]; then
        log "Deployment information:"
        echo "  Application URL: $(jq -r '.application_url.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")"
        echo "  Static Assets URL: $(jq -r '.static_assets_url.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")"
        echo "  CloudWatch Dashboard: https://$AWS_REGION.console.aws.amazon.com/cloudwatch/home?region=$AWS_REGION#dashboards:name=$(jq -r '.cloudwatch_dashboard_name.value' "$TERRAFORM_DIR/$ENVIRONMENT-outputs.json")"
    fi
}

# Run main function
main "$@"