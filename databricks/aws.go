package databricks

import (
	"github.com/xinsnake/databricks-sdk-golang"
	dbAws "github.com/xinsnake/databricks-sdk-golang/aws"
)

func NewAWSClient(host, token string) *dbAws.DBClient {
	o := databricks.DBClientOption{
		Host:  host,
		Token: token,
	}
	c := &dbAws.DBClient{
		Option: o,
	}
	c.Init(o)

	return c
}
