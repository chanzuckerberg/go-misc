package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	dataMessage = "DATA_MESSAGE"
)

type logStatus string

const (
	logStatusOK       logStatus = "OK"
	logStatusNoData   logStatus = "NODATA"
	logStatusSkipData logStatus = "SKIPDATA"
)

type augmentedLogEvent struct {
	Owner     string `json:"owner,omitempty"`
	LogGroup  string `json:"logGroup,omitempty"`
	LogStream string `json:"logStream,omitempty"`

	// From cloudwatch
	ID        string `json:"id,omitempty"`
	Timestamp int    `json:"timestamp,omitempty"`

	// Flow logs https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs.html
	Version     int    `json:"version,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	InterfaceID string `json:"interface_id,omitempty"`

	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`

	LogStatus logStatus `json:"log_status,omitempty"`

	// optional
	SourceAddress      *string `json:"source_address,omitempty"`
	DestinationAddress *string `json:"destination_address,omitempty"`
	SourcePort         *int    `json:"source_port,omitempty"`
	DestinationPort    *int    `json:"destination_port,omitempty"`
	Protocol           *int    `json:"protocol,omitempty"`
	Packets            *int    `json:"packets,omitempty"`
	Bytes              *int    `json:"bytes,omitempty"`
	Action             *string `json:"action,omitempty"`
}

// populate parses according to the Apache common format
// https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs.html
func (al *augmentedLogEvent) populate(message string) (err error) {
	split := strings.Split(message, " ")

	if len(split) != 14 {
		return errors.New("Malformed message")
	}

	al.Version, err = parseAsInt(split[0])
	if err != nil {
		return errors.Wrap(err, "Could not parse version")
	}
	al.AccountID = split[1]
	al.InterfaceID = split[2]

	startTime, err := parseAsInt(split[10])
	if err != nil {
		return errors.Wrap(err, "Could not parse start time")
	}
	al.StartTime = parseTime(startTime)

	endTime, err := parseAsInt(split[11])
	if err != nil {
		return errors.Wrap(err, "Could not parse end time")
	}
	al.EndTime = parseTime(endTime)

	al.LogStatus = logStatus(split[13])
	// End here if we don't have more data
	if al.LogStatus != logStatusOK {
		return nil
	}

	al.SourceAddress = optionalString(split[3])
	al.DestinationAddress = optionalString(split[4])

	al.SourcePort = parseAsOptionalInt(split[5])
	al.DestinationPort = parseAsOptionalInt(split[6])

	al.Protocol = parseAsOptionalInt(split[7])
	al.Packets = parseAsOptionalInt(split[8])
	al.Bytes = parseAsOptionalInt(split[9])
	al.Action = optionalString(split[12])
	return nil
}

func parseTime(t int) string {
	return time.Unix(int64(t), 0).Format(time.RFC3339)
}

func parseAsInt(val string) (int, error) {
	i, err := strconv.Atoi(val)
	return i, errors.Wrapf(err, "Could not parse %s as int", val)
}

func parseAsOptionalInt(val string) *int {
	i, err := parseAsInt(val)
	if err != nil {
		logrus.Error()
		return nil
	}
	return optionalInt(i)
}

func optionalString(s string) *string {
	return &s
}

func optionalInt(i int) *int {
	return &i
}

func processRecord(record events.KinesisFirehoseEventRecord) (response events.KinesisFirehoseResponseRecord, err error) {
	response.Data = record.Data
	response.RecordID = record.RecordID

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

	compressedMessages := bytes.NewBuffer(nil)
	messages := gzip.NewWriter(compressedMessages)
	b := []byte{}

	for _, logEvent := range parsed.LogEvents {
		augmented := augmentedLogEvent{}
		augmented.Owner = parsed.Owner
		augmented.LogGroup = parsed.LogGroup
		augmented.LogStream = parsed.LogStream

		err = augmented.populate(logEvent.Message)
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
		b, err = json.Marshal(augmented)
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
		_, err = messages.Write(b)
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}

		_, err = messages.Write([]byte{byte('\n')}) // newline
		if err != nil {
			response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			return
		}
	}
	// Close the gzip compression
	err = messages.Close()
	if err != nil {
		logrus.Errorf("Error compressing message %s", err)
		response.Result = events.KinesisFirehoseTransformedStateProcessingFailed
		return
	}
	response.Data = compressedMessages.Bytes()
	response.Result = events.KinesisFirehoseTransformedStateOk
	return
}

// HandleRequest handles a kinesis event request
func HandleRequest(ctx context.Context, kinesisEvent events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	response := events.KinesisFirehoseResponse{
		Records: []events.KinesisFirehoseResponseRecord{},
	}
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
