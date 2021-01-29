package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

type sink func() error // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func source() error {
	// Generate keypair, write to snowflake

	// Returns the keypair
	return nil
} // TODO: figure out inputs & outputs (pub key, priv key, user & snowflake account)

func databricksSink() error {
	// Same signature as sink type

	// Connecting to databricks
	return nil
}

func Run(ctx context.Context) error {
	err := source()
	if err != nil {
		return err
	}

	sinkList := []sink{databricksSink}
	for _, sink := range sinkList {
		err := sink()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	logrus.Info("hi")

	err := Run(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
}
