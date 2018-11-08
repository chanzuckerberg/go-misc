package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

const (
	dataMessage = "DATA_MESSAGE"
)

func processRecord(record events.KinesisEventRecord) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(record.Kinesis.Data))
	if err != nil {
		return err
	}
	defer gr.Close()
	parsed := &events.CloudwatchLogsData{}
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, parsed)
	if err != nil {
		return err
	}

	if parsed.MessageType != dataMessage {
		logrus.WithField("message_type", parsed.MessageType).Info("Skipping message")
		return nil
	}

	for _, logEvent := range parsed.LogEvents {
		func(logEvent events.CloudwatchLogsLogEvent) {
			hnyEvent := libhoney.NewEvent()
			hnyEvent.Metadata = logEvent.ID
			defer hnyEvent.Send()
			defer logrus.Infof("Sending event with id %s", hnyEvent.Metadata)

			// TODO: figure this out - honeycomb discards invalid timestamps
			//       but does not error on them
			// hnyEvent.Timestamp = time.Unix(0, logEvent.Timestamp)
			hnyEvent.Timestamp = time.Now()
			msg := map[string]interface{}{}
			err = json.Unmarshal([]byte(logEvent.Message), msg)
			if err != nil {
				logrus.WithError(err).Warn("Error json.Unmarshal")
				hnyEvent.AddField("message", logEvent.Message)
			} else {
				hnyEvent.Add(msg)
			}
			hnyEvent.AddField("aws.cloudwatch.group", parsed.LogGroup)
			hnyEvent.AddField("aws.cloudwatch.stream", parsed.LogStream)
			hnyEvent.AddField("aws.cloudwatch.owner", parsed.Owner)
			hnyEvent.AddField("aws.kinesis.region", record.AwsRegion)
			hnyEvent.AddField("aws.kinesis.source", record.EventSource)
		}(logEvent)
	}
	return nil
}

// HandleRequest handles a kinesis event request
func HandleRequest(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	dataset := os.Getenv("HONEYCOMB_DATASET")
	logrus.Infof("Sending events to %s", dataset)
	err := libhoney.Init(libhoney.Config{
		WriteKey: os.Getenv("HONEYCOMB_WRITE_KEY"),
		Dataset:  dataset,
	})
	if err != nil {
		logrus.WithError(err).Error("Error configuring honeycomb")
		return err
	}
	defer libhoney.Close()
	defer libhoney.Flush()

	logrus.Info("Processing started")
	logrus.Infof("Received %d kinesis records", len(kinesisEvent.Records))
	for _, record := range kinesisEvent.Records {
		err := processRecord(record)
		if err != nil {
			logrus.WithError(err).Error("Error processing records")
			return err
		}
	}
	logrus.Info("Success")
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
