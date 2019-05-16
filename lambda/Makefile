clean-cloudwatch_to_s3:
	@rm cloudwatch_to_s3.zip cloudwatch_to_s3 handler || true

package-cloudwatch_to_s3: clean-cloudwatch_to_s3
	GOOS=linux CGO_ENABLED=0 go build -o handler s3/cloudwatch_to_s3.go
	zip cloudwatch_to_s3.zip handler
	rm handler

clean-cloudwatch_to_honeycomb:
	@rm cloudwatch_to_honeycomb.zip cloudwatch_to_honeycomb handler || true

package-cloudwatch_to_honeycomb: clean-cloudwatch_to_honeycomb
	GOOS=linux CGO_ENABLED=0 go build -o handler honeycomb/cloudwatch_to_honeycomb.go
	zip cloudwatch_to_honeycomb.zip handler
	rm handler

clean-flow_to_s3:
	@rm flow_to_s3.zip flow_to_s3 handler || true

package-flow_to_s3: clean-flow_to_s3
	GOOS=linux CGO_ENABLED=0 go build -o handler flow/s3.go
	zip flow_to_s3.zip handler
	rm handler
