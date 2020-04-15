package aws

// Add mocks as necessary
//go:generate -command mock mockgen -package mocks -destination
//go:generate mock mocks/ec2.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API
//go:generate mock mocks/asg.go github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface AutoScalingAPI
//go:generate mock mocks/firehose.go github.com/aws/aws-sdk-go/service/firehose/firehoseiface FirehoseAPI
//go:generate mock mocks/iam.go github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
//go:generate mock mocks/kms.go github.com/aws/aws-sdk-go/service/kms/kmsiface KMSAPI
//go:generate mock mocks/lambda.go github.com/aws/aws-sdk-go/service/lambda/lambdaiface LambdaAPI
//go:generate mock mocks/s3.go github.com/aws/aws-sdk-go/service/s3/s3iface S3API
//go:generate mock mocks/s3manager.go github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface DownloaderAPI
//go:generate mock mocks/secretsmanager.go github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface SecretsManagerAPI
//go:generate mock mocks/ssm.go github.com/aws/aws-sdk-go/service/ssm/ssmiface SSMAPI
//go:generate mock mocks/sts.go github.com/aws/aws-sdk-go/service/sts/stsiface STSAPI
//go:generate mock mocks/support.go github.com/aws/aws-sdk-go/service/support/supportiface SupportAPI
//go:generate mock mocks/organizations.go github.com/aws/aws-sdk-go/service/organizations/organizationsiface OrganizationsAPI
