package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type sink func() error // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func source(sourceDetails ...string) (map[string]string, error) {
	// Generate keypair, write to snowflake
	for _, source := range sourceDetails {
		fmt.Println(source)
	}
	// Returns the keypair
	return nil, nil
} // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func databricksSink(sinkInputs map[string]string) error {
	// Same signature as sink type

	// Connecting to databricks
	return nil
}

func Run(ctx context.Context) error {
	sourceOutputs, err := source("account1")
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
