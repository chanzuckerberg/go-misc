package snowflake

// TODO: make this specific to an account
type SnowflakeClientEnv struct {
	REGION   string
	PASSWORD string `required:"true"`
	USER     string `required:"true"`
	ROLE     string `required:"true"`
}

type SnowflakeAccount struct {
	AppID string
	Name  string
}

type snowflakeAcctName string
type oktaAppID string
