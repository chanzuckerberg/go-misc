package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
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
	a.Equal(out.Result, events.KinesisFirehoseTransformedStateOk)

	r := bufio.NewReader(bytes.NewBuffer(out.Data))

	l, _, err := r.ReadLine()
	a.Nil(err)

	log := &augmentedLogEvent{}
	err = json.Unmarshal(l, log)
	a.Nil(err)

	a.Equal(log.LogGroup, data.LogGroup)
	a.Equal(log.LogStream, data.LogStream)
	a.Equal(log.Owner, data.Owner)
}
