DATE=$(shell date +"%Y/%m/%d")
S3_PREFIX ?= go-misc/lambdas/${DATE}
S3_BUCKET ?= shared-infra-prod-assets
AWS_PROFILE ?= czi-si

publish: publish-cloudwatch_to_s3 publish-cloudwatch_to_honeycomb publish-flow_to_s3 publish-github_to_firehose publish-unfurl_cloudtrail publish-cloudwatch_eks_to_s3 package-rotator-snowflake

clean-cloudwatch_to_s3:
	$(call clean,cloudwatch_to_s3.zip)

package-cloudwatch_to_s3: clean-cloudwatch_to_s3
	$(call package,s3/cloudwatch_to_s3.go,cloudwatch_to_s3.zip)

publish-cloudwatch_to_s3: package-cloudwatch_to_s3
	$(call publish,cloudwatch_to_s3.zip)

clean-cloudwatch_eks_to_s3:
	$(call clean,cloudwatch_eks_to_s3.zip)

package-cloudwatch_eks_to_s3: clean-cloudwatch_eks_to_s3
	$(call package,eks_s3/cloudwatch_eks_to_s3.go,cloudwatch_eks_to_s3.zip)

publish-cloudwatch_eks_to_s3: package-cloudwatch_eks_to_s3
	$(call publish,cloudwatch_eks_to_s3.zip)

clean-cloudwatch_to_honeycomb:
	$(call clean,cloudwatch_to_honeycomb.zip)

package-cloudwatch_to_honeycomb: clean-cloudwatch_to_honeycomb
	$(call package,honeycomb/cloudwatch_to_honeycomb.go,cloudwatch_to_honeycomb.zip)

publish-cloudwatch_to_honeycomb: package-cloudwatch_to_honeycomb
	$(call publish,cloudwatch_to_honeycomb.zip)

clean-flow_to_s3:
	$(call clean,flow_to_s3.zip)

package-flow_to_s3: clean-flow_to_s3
	$(call package,flow/s3.go,flow_to_s3.zip)

publish-flow_to_s3: package-flow_to_s3
	$(call publish,flow_to_s3.zip)

clean-github_to_firehose:
	$(call clean,github_to_firehose.zip)

package-github_to_firehose: clean-github_to_firehose
	$(call package,github/archive/archive.go,github_to_firehose.zip)

publish-github_to_firehose: package-github_to_firehose
	$(call publish,github_to_firehose.zip)

clean-unfurl_cloudtrail:
	$(call clean, unfurl_cloudtrail.zip)

package-unfurl_cloudtrail: clean-unfurl_cloudtrail
	$(call package,s3/unfurl/cloudtrail/main.go,unfurl_cloudtrail.zip)

publish-unfurl_cloudtrail: package-unfurl_cloudtrail
	$(call publish,unfurl_cloudtrail.zip)

clean-tfe-runall:
	$(call clean, tfe-runall.zip)

package-tfe-runall: clean-tfe-runall
	$(call package,tfe-runall/main.go,tfe-runall.zip)

publish-tfe-runall: package-tfe-runall
	$(call publish,tfe-runall.zip)

clean-tfe-metrics:
	$(call clean, tfe-metrics.zip)

package-tfe-metrics: clean-tfe-metrics
	$(call package,tfe-metrics/main.go,tfe-metrics.zip)

publish-tfe-metrics: package-tfe-metrics
	$(call publish,tfe-metrics.zip)

clean-tfe-backup:
	$(call clean, tfe-backup.zip)

package-tfe-backup: clean-tfe-backup
	$(call package,tfe-backup/main.go,tfe-backup.zip)

publish-tfe-backup: package-tfe-backup
	$(call publish,tfe-backup.zip)

clean-rotator-snowflake:
	$(call clean, rotator-snowflake.zip)

package-rotator-snowflake: clean-rotator-snowflake
	$(call package,rotator-snowflake/main.go,rotator-snowflake.zip)

publish-rotator-snowflake: package-rotator-snowflake
	$(call publish,rotator-snowflake.zip)

run-rotator-snowflake:
	cd rotator-snowflake/ && go run .

test:
	go test ./...

define clean
	# Removing the handler file
	@rm $1 handler 2> /dev/null  || true
endef

define package
	# Making the handler file
	GOOS=linux CGO_ENABLED=0 go build -o handler $1
	zip $2 handler
	rm handler
endef

define publish
	aws s3 cp $1 s3://${S3_BUCKET}/${S3_PREFIX}/$1 --profile ${AWS_PROFILE}
endef
