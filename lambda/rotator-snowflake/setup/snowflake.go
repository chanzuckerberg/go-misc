package setup

import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
)

func Snowflake(snowflakeAcct string) (*sql.DB, error) {
	snowflakeEnv, err := loadSnowflakeClientEnv()
	if err != nil {
		return nil, err
	}

	cfg := snowflake.SnowflakeConfig{
		Account:  snowflakeAcct,
		User:     snowflakeEnv.USER,
		Role:     snowflakeEnv.ROLE,
		Region:   snowflakeEnv.REGION,
		Password: snowflakeEnv.PASSWORD,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)

	return sqlDB, errors.Wrap(err, "Unable to configure Snowflake DB")
}
