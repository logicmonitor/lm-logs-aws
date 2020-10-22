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
  description = "LogicMonitor Access Id"
}

variable "lm_access_key" {
  description = "LogicMonitor Access Key"
}

variable "lm_company_name" {
  description = "LogicMonitor Account Name"
}

# LogicMonitor Logs forwarder
resource "aws_cloudformation_stack" "lm_forwarder" {
  name         = "lm-forwarder"
  capabilities = ["CAPABILITY_IAM", "CAPABILITY_NAMED_IAM", "CAPABILITY_AUTO_EXPAND"]
  parameters   = {
    FunctionName              = "LMLogsForwarder"
    LMAccessId                = var.lm_access_id
    LMAccessKey               = var.lm_access_key
    LMCompanyName             = var.lm_company_name
    LMRegexScrub              = ""
    PermissionsBoundaryArn    = ""
  }
  template_url = "https://lm-logs-forwarder.s3.amazonaws.com/latest.yaml"
}
```
`terraform apply --var 'lm_access_id=<lm_access_id>' --var 'lm_access_key=<lm_access_key>' --var 'lm_company_name=<lm_company_name>'`

### Forwarding EC2 Instances logs
Forward EC2 logs to CloudWatch, using the [CloudWatch Logs Agent](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/QuickStartEC2Instance.html). 
**Note:** The logstream name typically defaults to the instance ID (this is expected by LogicMonitor).
After you have started recieving your EC2 logs in the desired log group:
1. Go to CloudWatch, select the desired log group of which you want to forward logs , under Actions > Create Lambda subscription filter
2. In Create Lambda subscription filter , select "Lambda Function" and choose "LMLogsForwarder" (or, whatever you named the Lambda function during stack creation) and click Start streaming.

### Forwarding S3 bucket access logs
To forward S3 bucket access logs to LogicMonitor: 
1. Under the bucket **Properties**, enable **Server access logging**.
   You will need to select a **Target bucket** where the access logs will be stored. If this target bucket doesn't exist, you need to create it. (This is different from the source bucket.)
2. Go to the target bucket, and under **Advanced settings > Events** add a notification for "All object create events". 
3. **Send to** "Lambda Function" and choose "LMLogsForwarder" (or, whatever you named the Lambda function during stack creation).

### Forwarding ELB access logs
To send ELB access logs to LogicMonitor:
1. In the EC2 navigation pane, choose Load Balancers and select your load balancer.
2. Under Attributes > Access logs, click "Configure access logs".
3. Select "Enable access logs" and specify the S3 bucket to store the logs. (You can create one, if it doesn't exist.)
4. Go to the S3 bucket (from Step 3), and under **Advanced settings > Events** add a notification for "All object create events". 
5. **Send to** "Lambda Function" and choose "LMLogsForwarder" (or, whatever you named the Lambda function during stack creation).

### Forwarding RDS logs
To send RDS logs to LogicMonitor, configure instance to send logs to cloudwatch, and create subscription filter to send logs to the LM log forwarder:
1. Follow instructions to send [standard RDS logs to cloudwatch](https://aws.amazon.com/blogs/database/monitor-amazon-rds-for-mysql-and-mariadb-logs-with-amazon-cloudwatch/) or  [enhanced RDS to cloudwatch](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.html)
2. Go to Cloudwatch, select the desired log group of which you want to forward logs , under Actions > Create Lambda subscription filter
3. In Create Lambda subscription filter , select "Lambda Function" and choose "LMLogsForwarder" (or, whatever you named the Lambda function during stack creation) and click Start streaming.

### Forwarding Lambda logs
To send Lambda logs to LogicMonitor, go to cloudwatch and find lambda's log group, and create subscription filter to send logs to the LM log forwarder:
1. Go to Cloudwatch, select the lambda's log group of which you want to forward logs , under Actions > Create Lambda subscription filter
2. In Create Lambda subscription filter , select "Lambda Function" and choose "LMLogsForwarder" (or, whatever you named the Lambda function during stack creation) and click Start streaming.