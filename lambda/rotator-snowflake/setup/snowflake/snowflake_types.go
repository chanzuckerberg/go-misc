package snowflake

type SnowflakeClientEnvironment struct {
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

type SnowflakeAccountEnvironment struct {
	OktaMap map[snowflakeAcctName]oktaAppID `required:"true"`
}
