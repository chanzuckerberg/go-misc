package databricks

import (
	"github.com/xinsnake/databricks-sdk-golang"
	DBAws "github.com/xinsnake/databricks-sdk-golang/aws"
)

func NewAWSClient(host, token string) *DBAws.DBClient {
	o := databricks.DBClientOption{
		Host:  host,
		Token: token,
	}
	c := &DBAws.DBClient{
		Option: o,
	}
	c.Init(o)

	return c
}
