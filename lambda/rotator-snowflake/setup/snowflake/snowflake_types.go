package snowflake

import "database/sql"

// TODO: make this specific to an account
type SnowflakeClientEnv struct {
	ACCOUNT  string `required:"true"`
	APP_ID   string `required:"true"`
	PASSWORD string `required:"true"`
	USER     string `required:"true"`
	ROLE     string `required:"true"`
	REGION   string
}

type SnowflakeAccount struct {
	AppID string
	Name  string
	DB    *sql.DB
}

type Accounts struct {
	ACCOUNTS []string `required:"true"`
}
