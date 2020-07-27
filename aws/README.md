# lm-logs-aws-integration (beta)
Cloud formation template to push logs from AWS to Logic Monitor. You need to have basic infrastructure setup for that you need to deploy stack
To deploy you need.
* LM Access ID
* LM Access Key
* KM Account Name

### Deploying lambda using CloudFormation
[![Launch Stack](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?stackName=lm-forwarder&templateURL=https://lm-logs-forwarder.s3.amazonaws.com/latest.yaml)

### Forwarding EC2 Instances logs

Use this guide the set that up: [Configure the CloudWatch Logs Agent on EC2](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/QuickStartEC2Instance.html)
and send logs to `/aws/lambda/lm` log group.

### Forwarding S3 bucket access logs
By default, logging is disabled. When logging is enabled, logs are saved to a bucket in the same AWS Region as the source bucket.
To enable access logging and send it to Logic Monitor, you must do the following:
* Create a **target bucket** where you want to store access logs.
* Go to the properties tab of the **source bucket**, click on **Server Access logs** and enable logging and select the **target bucket**.
* Configure **trigger** on the **target bucket** which contains your access logs, and change the event type to **Object Created (All)** and select lam ** LMLogsForwarder**

### Forwarding ELB access logs
* In the navigation pane, choose Load Balancers.Select your load balancer.
* On the Description tab, choose Edit attributes.
* On the Edit load balancer attributes page, do the following:
* For Access logs, select Enable and type the name of the bucket in which logs with be dumped.
* Configure **trigger** on the bucket which contains your access logs, and change the event type to **Object Created (All)** and select lam ** LMLogsForwarder**