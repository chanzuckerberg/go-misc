package main

import (
	"context"
	"flag"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/rotate"
	"github.com/sirupsen/logrus"
)

var localFlag = flag.Bool("local", false, "Whether this lambda should be run locally")

func Run(ctx context.Context) error {
	return rotate.Rotate(ctx)
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	flag.Parse()
	logrus.Debugf("localFlag: %t", *localFlag)

	// local-mode for lambda
	if *localFlag {
		err := Run(context.Background())
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		lambda.Start(Run)
	}
}
