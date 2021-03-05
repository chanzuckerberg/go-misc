package databricks

import (
	"github.com/chanzuckerberg/go-misc/databricks"
)

func Databricks() (DatabricksConnection, error) {
	databricksEnv, err := loadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return DatabricksConnection(dbClient), nil
}
