package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	dataMessage = "DATA_MESSAGE"
)

/*
   "'use strict';",
   "const zlib = require('zlib');",
   "function transformLogEvent(logEvent) {",
   "       return Promise.resolve(`${logEvent.message}\n`);",
   "}",
   "exports.handler = (event, context, callback) => {",
   "    Promise.all(event.records.map(r => {",
   "        const buffer = new Buffer(r.data, 'base64');",
   "        const decompressed = zlib.gunzipSync(buffer);",
   "        const data = JSON.parse(decompressed);",
   "        if (data.messageType !== 'DATA_MESSAGE') {",
   "            return Promise.resolve({",
   "                recordId: r.recordId,",
   "                result: 'ProcessingFailed',",
   "            });",
   "         } else {",
   "            const promises = data.logEvents.map(transformLogEvent);",
   "            return Promise.all(promises).then(transformed => {",
   "                const payload = transformed.reduce((a, v) => a + v, '');",
   "                const encoded = new Buffer(payload).toString('base64');",
   "                console.log('---------------payloadv2:'+JSON.stringify(payload, null, 2));",
   "                return {",
   "                    recordId: r.recordId,",
   "                    result: 'Ok',",
   "                    data: encoded,",
   "                };",
   "           });",
   "        }",
   "    })).then(recs => callback(null, { records: recs }));",
     "};"
*/

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
			return response, err
		}
		response.Records = append(response.Records, record)
	}
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
