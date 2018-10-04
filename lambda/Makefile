clean-cloudwatch_to_s3:
	rm cloudwatch_to_s3.zip cloudwatch_to_s3

package-cloudwatch_to_s3: clean-cloudwatch_to_s3
	GOOS=linux CGO_ENABLED=0 go build cloudwatch_to_s3.go
	zip cloudwatch_to_s3.zip ./cloudwatch_to_s3
	rm cloudwatch_to_s3
