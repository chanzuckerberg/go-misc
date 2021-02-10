package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func Run(ctx context.Context) error {
	return rotate(ctx)
}

func main() {
	logrus.Info("hi")

	if err := Run(context.Background()); err != nil {
		logrus.Fatal(err)
	}
}
