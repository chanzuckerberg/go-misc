# tfe-runall

This lambda takes the name of a tfe org and runs a plan in each workspace.

Passing the `-local` flag will run the command locally, without any lambda infrastructure.

Required environment variales-

* TFE_ORG
* TFE_TOKEN_SECRET_ARN (arn of an aws secretsmanager secret which includes a TFE team token)

Optional environment variables–

* TFE_ADDRESS (if not using TFCloud)
* any other [go-tfe](https://github.com/hashicorp/go-tfe) env variables

Example run for local dev-

```shell
export TFE_ORG='tf-scratch'
export TFE_TOKEN_SECRET_ARN='arn:aws:secretsmanager:us-west-2:1234567890:secret:foobar'
go run ./lambda/tfe-runall -- -local
```
