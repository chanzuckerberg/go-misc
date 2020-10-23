# tfe-metrics

## deploy

`make deploy-dev`

or

`make deploy-prod`


## logs

install https://github.com/jorgebastida/awslogs and

```
AWS_PROFILE=czi-tfe AWS_REGION=us-west-2 awslogs get /aws/lambda/tfe-prod-metrics -w
```
