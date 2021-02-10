package main

import (
	"context"

	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/rotate"
	"github.com/sirupsen/logrus"
)

func Run(ctx context.Context) error {
	return rotate.Rotate(ctx)
}

func main() {
	if err := Run(context.Background()); err != nil {
		logrus.Fatal(err)
	}
}
