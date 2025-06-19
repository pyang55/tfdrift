resource "aws_sqs_queue" "demo_queue" {
  name                      = var.queue_name
  delay_seconds             = 0
  message_retention_seconds = 86400

  tags = {
    Environment = "De"
    Owner       = "Terraform"
  }
}
