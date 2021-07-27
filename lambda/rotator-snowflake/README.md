# Guide to rotator-snowflake
This is a package that connects to Okta, Snowflake, and Databricks in order to 

## Requirements
### Okta
You'll need to set up a keypair to use with an Okta OAuth app so you can talk to both Snowflake and Databricks apps.

If you use Terraform, this block will help you set up the Okta app:
```hcl
resource "okta_app_oauth" "<app name here>" {
    label         = "<app label>"
    type          = "service"
    grant_types   = ["client_credentials"]
    redirect_uris = ["https://<databricks domain>/oauth2/callback"]
    login_uri     = "https://<databricks domain>"
    response_types = ["token"]
    token_endpoint_auth_method = "private_key_jwt
    jwks {
        kid = ""
        kty = ""
        e   = ""
        n   = ""
    }
}
```
To generate the jwks values (save the private key separately):
```bash
$ aws-oidc rsa-keygen 
{"use":"...","kty":"...","kid":"...","alg":"...","n":"...","e":"..."}
```

You'll also need the app's client ID (refer to the Terraform code) and the Okta Identity Provider URL (some `domain.okta.com`) 

Assumption: We're also assuming the Databricks and Snowflake apps are both hosted on the same Okta provider, so you'll need their client IDs. 

### Exposed Environment Variables
#### Okta
You'll need the OAuth app's Identity Provider URL and Client ID. 

#### Databricks
You'll need the Databricks host and its Okta Client ID. 

#### Snowflake
You'll need a Snowflake user. To configure it, you'll need the user's account name, role name, and username. You'll also need a mapping between the account name and its Okta Client ID. Defining the map would look like:
```
SNOWFLAKE_OKTAMAP=accountName1:clientID1,accountName2:clientID2
```

### Sensitive Environment Variables
We save these sensitive variables in specific paths in the AWS Parameter store so the Environment Variables won't be exposed in the Lambda configuration.

Each PARAM_STORE_SERVICE env variable helps us locate the service-specific sensitive variables, like this: `/<okta, databricks, or snowflake service path>/<sensitivesecretname>`
* OKTA_PARAM_STORE_SERVICE path:  
  * okta_private_key: this is where we load the okta private key from the Okta section  
* DATABRICKS_PARAM_STORE_SERVICE path:  
  * databricks_token: API token used for authenticating into Databricks  
* SNOWFLAKE_accountname_PARAM_STORE_SERVICE (one for each snowflake account you want to configure):  
  * snowflake_accountname_password: Snowflake user's password

### Summary of Configuration Requirements
```
# Exposed Okta Environment
OKTA_ORG_URL
OKTA_CLIENT_ID 

# Exposed Snowflake Environment
SNOWFLAKE_OKTAMAP
SNOWFLAKE_<accountName1>_NAME
SNOWFLAKE_<accountName1>_ROLE 
SNOWFLAKE_<accountName1>_USER 
<continue name, role, user for accountName2 and beyond>

# Exposed Databricks Environment
DATABRICKS_HOST
DATABRICKS_APP_ID

# Paths to sensitive secrets in AWS Secrets Manager, which is in Systems Manager (SSM)
OKTA_PARAM_STORE_SERVICE (contains okta_private_key)
DATABRICKS_PARAM_STORE_SERVICE (contains DATABRICKS_PARAM_STORE_SERVICE)
SNOWFLAKE_<accountName1>_PARAM_STORE_SERVICE (contains snowflake_accountName1_password)
```

## Run the code
### Locally
```bash
$ cd go-misc/lambda/rotator-snowflake
rotator-snowflake $  export ...
rotator-snowflake $  go run main.go -local true
```
Locally using aws-oidc:
```bash
$ cd go-misc/lambda/rotator-snowflake
rotator-snowflake $ AWS_PROFILE=<profile> aws-oidc exec <AWS_PROFILE> -- go run main.go -local true
```
### With AWS Lambda
1. Publish the package
```bash
$ cd go-misc/lambda/
lambda $ make publish-rotator-snowflake
```
2. Reference S3 handler file in your Lambda code. We suggest using our [Lambda Function](https://github.com/chanzuckerberg/cztack/tree/main/aws-lambda-function) and [Secrets Reader](https://github.com/chanzuckerberg/cztack/tree/main/aws-iam-secrets-reader-policy) Terraform modules.
