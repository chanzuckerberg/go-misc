package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-lambda-go/events"
)

func TestS3ProcessRecord(t *testing.T) {
	a := assert.New(t)

	in := events.KinesisFirehoseEventRecord{
		RecordID: "AAA",
		Data:     []byte{},
	}

	data := events.CloudwatchLogsData{
		Owner:       "an owner",
		LogGroup:    "a log group",
		LogStream:   "a log stream",
		MessageType: dataMessage,
		LogEvents: []events.CloudwatchLogsLogEvent{
			{
				ID:        "asfd",
				Message:   "2 123456789010 eni-abc123de 172.31.16.139 172.31.16.21 20641 22 6 20 4249 1418530010 1418530070 ACCEPT OK",
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
	a.Equal(out.Result, events.KinesisFirehoseTransformedStateOk)
}
