clean-cloudwatch_to_s3:
	@rm cloudwatch_to_s3.zip handler || true

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
	@rm flow_to_s3.zip handler || true

package-flow_to_s3: clean-flow_to_s3
	GOOS=linux CGO_ENABLED=0 go build -o handler flow/s3.go
	zip flow_to_s3.zip handler
	rm handler

clean-github_to_firehose:
	@rm github_to_firehose.zip handler || true

package-github_to_firehose: clean-github_to_firehose
	GOOS=linux CGO_ENABLED=0 go build -o handler github/archive/archive.go
	zip github_to_firehose.zip handler
	rm handler

clean-unfurl_cloudtrail:
	@rm unfurl_cloudtrail.zip handler || true

package-unfurl_cloudtrail: clean-unfurl_cloudtrail
	GOOS=linux CGO_ENABLED=0 go build -o handler s3/unfurl/cloudtrail/main.go
	zip unfurl_cloudtrail.zip handler
	rm handler
