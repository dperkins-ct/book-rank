# Route 53 DNS Configuration

# Route 53 Hosted Zone (assumes you have your domain)
data "aws_route53_zone" "main" {
  count        = var.environment == "production" ? 1 : 0
  name         = var.domain_name
  private_zone = false
}

# ACM Certificate for ALB (regional)
resource "aws_acm_certificate" "main" {
  domain_name               = local.domain_name
  subject_alternative_names = var.environment == "production" ? ["*.${var.domain_name}"] : []
  validation_method         = "DNS"

  tags = local.common_tags

  lifecycle {
    create_before_destroy = true
  }
}

# ACM Certificate for CloudFront (must be in us-east-1)
resource "aws_acm_certificate" "cloudfront" {
  count    = var.environment == "production" ? 1 : 0
  provider = aws.us_east_1

  domain_name       = "static.${var.domain_name}"
  validation_method = "DNS"

  tags = local.common_tags

  lifecycle {
    create_before_destroy = true
  }
}

# ACM Certificate Validation Records
resource "aws_route53_record" "cert_validation" {
  for_each = var.environment == "production" ? {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.main[0].zone_id
}

resource "aws_route53_record" "cloudfront_cert_validation" {
  for_each = var.environment == "production" ? {
    for dvo in aws_acm_certificate.cloudfront[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.main[0].zone_id
}

# ACM Certificate Validation
resource "aws_acm_certificate_validation" "main" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = var.environment == "production" ? [for record in aws_route53_record.cert_validation : record.fqdn] : null

  timeouts {
    create = "5m"
  }
}

resource "aws_acm_certificate_validation" "cloudfront" {
  count    = var.environment == "production" ? 1 : 0
  provider = aws.us_east_1

  certificate_arn         = aws_acm_certificate.cloudfront[0].arn
  validation_record_fqdns = [for record in aws_route53_record.cloudfront_cert_validation : record.fqdn]

  timeouts {
    create = "5m"
  }
}

# Route 53 Records
resource "aws_route53_record" "main" {
  count   = var.environment == "production" ? 1 : 0
  zone_id = data.aws_route53_zone.main[0].zone_id
  name    = local.domain_name
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "static" {
  count   = var.environment == "production" ? 1 : 0
  zone_id = data.aws_route53_zone.main[0].zone_id
  name    = "static.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.static_assets.domain_name
    zone_id                = aws_cloudfront_distribution.static_assets.hosted_zone_id
    evaluate_target_health = false
  }
}

# Health Check for production
resource "aws_route53_health_check" "main" {
  count                            = var.environment == "production" ? 1 : 0
  fqdn                            = local.domain_name
  port                            = 443
  type                            = "HTTPS"
  resource_path                   = "/health"
  failure_threshold               = 3
  request_interval                = 30
  cloudwatch_alarm_region         = var.aws_region
  cloudwatch_alarm_name           = aws_cloudwatch_metric_alarm.health_check[0].alarm_name
  insufficient_data_health_status = "Failure"

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-health-check"
  })
}

# CloudWatch Alarm for Health Check
resource "aws_cloudwatch_metric_alarm" "health_check" {
  count               = var.environment == "production" ? 1 : 0
  alarm_name          = "${local.name_prefix}-health-check-failed"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "1"
  alarm_description   = "This metric monitors the health check status"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "breaching"

  dimensions = {
    HealthCheckId = aws_route53_health_check.main[0].id
  }

  tags = local.common_tags
}