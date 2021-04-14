package snowflake

import "database/sql"

// envconfig that's prefixed by the snowflake account name
type SnowflakeClientEnv struct {
	NAME                string `required:"true"`
	USER                string `required:"true"`
	ROLE                string `required:"true"`
	PARAM_STORE_SERVICE string `required:"true"`
	REGION              string
}

type Account struct {
	AppID string
	Name  string
	DB    *sql.DB
}

type Accounts struct {
	OKTAMAP map[string]string `required:"true"`
}
