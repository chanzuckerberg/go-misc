package snowflake

type SnowflakeClientEnvironment struct {
	ACCOUNT  string `required:"true"`
	REGION   string
	PASSWORD string `required:"true"`
	USER     string `required:"true"`
	ROLE     string `required:"true"`
}
