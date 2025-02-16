# EC2 uptime maestro lambda

## Prerequisites

EC2 instances need to have the tag: `ec2maestro:yes`

## Build binary

Run: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go`

## Links

- Used in open tofu module: https://github.com/jzoric/ec2-uptime-maestro-tofu
- Used in AWS CDK: https://github.com/jzoric/ec2-uptime-maestro-cdk
