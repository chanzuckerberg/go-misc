package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestS3ProcessRecordDropped(t *testing.T) {
	a := assert.New(t)

	in := events.KinesisFirehoseEventRecord{
		RecordID: "record1",
		Data:     []byte{},
	}

	data := events.CloudwatchLogsData{
		Owner:       "owner1",
		LogGroup:    "loggroup",
		MessageType: dataMessage,
		LogEvents: []events.CloudwatchLogsLogEvent{
			{
				ID:        "asfd",
				Message:   "asfd",
				Timestamp: 123123,
			},
		},
	}

	b, err := json.Marshal(data)
	a.Nil(err)

	compressed := bytes.NewBuffer(nil)
	g := gzip.NewWriter(compressed)
	_, err = g.Write(b)
	a.Nil(err)
	err = g.Close()
	a.Nil(err)
	in.Data = compressed.Bytes()

	out, err := processRecord(in)
	a.Nil(err)

	a.Equal(out.RecordID, in.RecordID)
	a.Equal(out.Result, events.KinesisFirehoseTransformedStateDropped)
}
