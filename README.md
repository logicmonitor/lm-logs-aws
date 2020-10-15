# lm-logs-aws-integration (beta)
This integration provides a CloudFormation template to forward logs from AWS CloudWatch to LogicMonitor.

This CloudFormation template only deploys a log forwarder (lambda function) subscribed to a specific CloudWatch logs group for LogicMonitor. 
Forwarding logs from individual AWS services, such as EC2, S3, or ELB, should be configured separately.

You will need to supply the following LogicMonitor credentials when configuring the CloudFormation stack:
* LM Access ID
* LM Access Key
* LM Account Name

### Deploying lambda using CloudFormation
[![Launch Stack](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?stackName=lm-forwarder&templateURL=https://lm-logs-forwarder.s3.amazonaws.com/latest.yaml)

### Deploying lambda using Terraform

**Sample configuration**
```tf
variable "lm_access_id" {
  description = "Logic Monitor Access Id"
}

variable "lm_access_key" {
  description = "Logic Monitor Access Key"
}

variable "host_url" {
  description = "Host Url"
}

variable "log_group_name" {
  description = "Cloudwatch log group name"
}

# Logic Monitor Logs forwarder
resource "aws_cloudformation_stack" "lm_forwarder" {
  name         = "lm-forwarder"
  capabilities = ["CAPABILITY_IAM", "CAPABILITY_NAMED_IAM", "CAPABILITY_AUTO_EXPAND"]
  parameters   = {
    FunctionName              = "LMLogsForwarder"
    LMAccessId                = var.lm_access_id
    LMAccessKey               = var.lm_access_key
    LMIngestEndpoint          = var.host_url
    LMRegexScrub              = ""
    LogGroupName              = var.log_group_name
    LogGroupRetentionInDays   = 90
    PermissionsBoundaryArn    = ""
  }
  template_url = "https://lm-logs-forwarder.s3.amazonaws.com/latest.yaml"
}
```
`terraform apply --var 'lm_access_id=<lm_access_id>' --var 'lm_access_key=<lm_access_key>' --var 'host_url=<host_url>' --var 'log_group_name=<log_group_name>'`

### Forwarding EC2 Instances logs

There are several ways to forward EC2 logs, including using the [CloudWatch Logs Agent](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/QuickStartEC2Instance.html), Fluentd, and more. 
The logstream name typically defaults to the instance ID (this is expected by LogicMonitor).

### Forwarding S3 bucket access logs
To forward S3 access logs to LogicMonitor, make sure that logging is enabled and that you are sending events to the 
Lambda function forwarding logs to LogicMonitor. 

1. Enable logging for the bucket that contains the access logs you want to forward.
2. Under Advanced settings > Events, add a notification for “All object create events”. 
3. Send to “Lambda Function” and choose “LMLogsForwarder” (or, whatever you named the Lambda function during stack creation).

### Forwarding ELB access logs
To send ELB access logs to LogicMonitor, enable access logging to an S3 bucket and configure the events to send to the LM log forwarder:

1. In the EC2 navigation pane, choose Load Balancers and select your load balancer.
2. Under Attributes > Access logs, click “Configure access logs”.
3. Select “Enable access logs” and specify the S3 bucket to store the logs. (You can create one, if it doesn’t exist.)
4. In S3, configure the bucket to forward events to the Lambda Function.

### Forwarding RDS logs
To send RDS logs to LogicMonitor, configure instance to send logs to cloudwatch, and create subscription filter to send logs to the LM log forwarder:

1. Follow instructions to send [standard RDS logs to cloudwatch](https://aws.amazon.com/blogs/database/monitor-amazon-rds-for-mysql-and-mariadb-logs-with-amazon-cloudwatch/) or  [enhanced RDS to cloudwatch](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.html)
2. Go to Cloudwatch, select the desired log group of which you want to forward logs , under Actions > Create Lambda subscription filter
3. In Create Lambda subscription filter , select “Lambda Function” and choose “LMLogsForwarder” (or, whatever you named the Lambda function during stack creation) and click Start streaming.

### Forwarding Lambda logs
To send Lambda logs to LogicMonitor, go to cloudwatch and find lambda's log group, and create subscription filter to send logs to the LM log forwarder:

1. Go to Cloudwatch, select the lambda's log group of which you want to forward logs , under Actions > Create Lambda subscription filter
2. In Create Lambda subscription filter , select “Lambda Function” and choose “LMLogsForwarder” (or, whatever you named the Lambda function during stack creation) and click Start streaming.