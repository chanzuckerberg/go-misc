package snowflake

// func LoadSnowflakeAccounts(accountMap map[string]string, secrets setup.SecretStore) ([]*Account, error) {
// 	snowflakeErrs := &multierror.Error{}
// 	acctList := []*Account{}

// 	for acctName, snowflakeAppID := range accountMap {
// 		// If acctName has "okta" or "databricks" in the name, print a warning for possible name collision
// 		oktaCollision := strings.Contains(acctName, "okta")
// 		if oktaCollision {
// 			logrus.Warnf("Snowflake Account %s will likely collide with okta Environment Variables", acctName)
// 		}

// 		databricksCollision := strings.Contains(acctName, "databricks")
// 		if databricksCollision {
// 			logrus.Warnf("Snowflake Account %s will likely collide with databricks Environment Variables", acctName)
// 		}

// 		snowflakeEnv := &SnowflakeClientEnv{}

// 		err := envconfig.Process(acctName, snowflakeEnv)
// 		if err != nil {
// 			snowflakeErrs = multierror.Append(snowflakeErrs, errors.Wrap(err, "Error processing Snowflake environment variables"))
// 		}

// 		sqlDB, err := ConfigureConnection(snowflakeEnv, secrets)
// 		if err != nil {
// 			snowflakeErrs = multierror.Append(snowflakeErrs, err)

// 			continue
// 		}

// 		acctList = append(acctList, &Account{
// 			AppID: snowflakeAppID,
// 			Name:  snowflakeEnv.NAME,
// 			DB:    sqlDB,
// 		})
// 	}

// 	return acctList, snowflakeErrs.ErrorOrNil()
// }
