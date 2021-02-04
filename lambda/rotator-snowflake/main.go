package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/sirupsen/logrus"
	databricks "github.com/xinsnake/databricks-sdk-golang"
)

type sink func() error // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func configureSnowflake(sourceDetails ...snowflake.SnowflakeConfig) (map[string]string, error) {
	// Generate keypair, write to snowflake
	for _, source := range sourceDetails {
		fmt.Println(source)
		// Connect to snowflake

		// Execute snowflake query

		//
	}

	// Returns the keypair
	return nil, nil
} // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func databricksSink(sinkInputs map[string]string) error {
	// Connect to Databricks
	var o databricks.DBClientOption
	o.Host = os.Getenv("DATABRICKS_HOST")
	o.Token = os.Getenv("DATABRICKS_TOKEN")

	return nil
}

func Run(ctx context.Context) error {
	// read snowflake YAML file? or command line arguments?
	inputConfiguration := snowflake.SnowflakeConfig{}
	sourceOutputs, err := configureSnowflake(inputConfiguration)
	if err != nil {
		return err
	}

	err = databricksSink(sourceOutputs)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	logrus.Info("hi")

	if err := Run(context.Background()); err != nil {
		logrus.Fatal(err)
	}
}
