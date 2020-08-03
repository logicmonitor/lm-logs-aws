# lm-logs-aws-integration (beta)
This integration provides a CloudFormation template to forward logs from AWS CloudWatch to LogicMonitor.

You will need to supply the following LogicMonitor credentials when configuring the CloudFormation stack:
* LM Access ID
* LM Access Key
* LM Account Name

### Deploying lambda using CloudFormation
[![Launch Stack](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?stackName=lm-forwarder&templateURL=https://lm-logs-forwarder.s3.amazonaws.com/latest.yaml)

### Forwarding EC2 Instances logs

There are several ways to forward EC2 logs, including using the [CloudWatch Logs Agent](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/QuickStartEC2Instance.html), Fluentd, and more. 
* Send the logs to an existing log group, such as `/aws/lambda/lm`, to LogicMonitor.
* The logstream name typically defaults to the instance id.

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
