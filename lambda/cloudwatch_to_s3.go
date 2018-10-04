package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

const (
	dataMessage = "DATA_MESSAGE"
)

func processRecord(record events.KinesisFirehoseEventRecord) (response events.KinesisFirehoseResponseRecord, err error) {
	response.Data = record.Data
	response.RecordID = record.RecordID
	logrus.Infof("Processing record %s", record.RecordID)

	gr, err := gzip.NewReader(bytes.NewBuffer(record.Data))
	if err != nil {
		response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
		return
	}
	defer gr.Close()

	parsed := &events.CloudwatchLogsData{}
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
		return
	}
	err = json.Unmarshal(data, parsed)
	if err != nil {
		response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
		return
	}
	if parsed.MessageType != dataMessage {
		response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
		return
	}
	messages := bytes.NewBuffer(nil)
	b := []byte{}
	for _, logEvent := range parsed.LogEvents {
		b, err = json.Marshal(logEvent)
		if err != nil {
			response.Data = messages.Bytes()
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
		_, err = messages.Write(b)
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
		_, err = messages.WriteRune('\n')
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
	}
	response.Data = messages.Bytes()
	response.Result = events.KinesisFirehoseTransformedStateOk
	logrus.Infof("Successfully parsed %d messages in recordID %s", len(parsed.LogEvents), record.RecordID)
	return
}

// HandleRequest handles a kinesis event request
func HandleRequest(ctx context.Context, kinesisEvent events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	response := events.KinesisFirehoseResponse{
		Records: []events.KinesisFirehoseResponseRecord{},
	}
	logrus.Infof("Received %d kinesis records", len(kinesisEvent.Records))
	for _, record := range kinesisEvent.Records {
		record, err := processRecord(record)
		if err != nil {
			logrus.WithError(err).Error("Error processing records")
			return response, err
		}
		response.Records = append(response.Records, record)
	}
	logrus.Info("Success")
	return response, nil
}

func main() {
	logrus.Info("Processing started")
	lambda.Start(HandleRequest)
}
