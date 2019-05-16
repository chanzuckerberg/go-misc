clean-cloudwatch_to_s3:
	@rm cloudwatch_to_s3.zip cloudwatch_to_s3 || true

package-cloudwatch_to_s3: clean-cloudwatch_to_s3
	GOOS=linux CGO_ENABLED=0 go build -o cloudwatch_to_s3 s3/cloudwatch_to_s3.go
	zip cloudwatch_to_s3.zip ./cloudwatch_to_s3
	rm cloudwatch_to_s3

clean-cloudwatch_to_honeycomb:
	@rm cloudwatch_to_honeycomb.zip cloudwatch_to_honeycomb || true

package-cloudwatch_to_honeycomb: clean-cloudwatch_to_honeycomb
	GOOS=linux CGO_ENABLED=0 go build -o cloudwatch_to_honeycomb honeycomb/cloudwatch_to_honeycomb.go
	zip cloudwatch_to_honeycomb.zip ./cloudwatch_to_honeycomb
	rm cloudwatch_to_honeycomb

clean-flow_to_s3:
	@rm flow_to_s3.zip flow_to_s3 || true

package-flow_to_s3: clean-flow_to_s3
	GOOS=linux CGO_ENABLED=0 go build -o flow_to_s3 flow/s3.go
	zip flow_to_s3.zip flow_to_s3
	rm flow_to_s3
