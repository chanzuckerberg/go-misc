package state

// this is a derivative of https://github.com/honeycombio/honeyaws/blob/master/state/state.go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	stateFileFormat = "%s-state.json"
)

// Stater lets us gain insight into the current state of object processing. It
// could be backed by the local filesystem, cloud abstractions such as
// DynamoDB, consistent value stores like etcd, etc.
type Stater interface {
	// SetProcessed indicates that downloading, processing, and sending the
	// object to Honeycomb has been completed successfully.
	SetProcessed(object string) error

	IsProcessed(object string) (bool, error)
}

// Used to communicate between the various pieces which are relying on state
// information.
type DownloadedObject struct {
	Object, Filename string
}

type DynamoDBStater struct {
	session   *session.Session
	svc       *dynamodb.DynamoDB
	TableName string
}

func NewDynamoDBStater(session *session.Session, tableName string) (*DynamoDBStater, error) {
	svc := dynamodb.New(session)
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}
	_, err := svc.DescribeTable(input)

	stater := &DynamoDBStater{
		session:   session,
		TableName: tableName,
		svc:       svc,
	}

	if err != nil {
		// For some reason, we cannot write to
		// the table or access it
		return stater, err
	}

	return stater, nil
}

// Used for unmarshaling and adding objects to DynamoDB
type Record struct {
	S3Object string
	Time     time.Time
}

func (d *DynamoDBStater) SetProcessed(s3object string) error {

	log.Printf("set processed %s", s3object)
	svc := dynamodb.New(d.session)

	objMap := Record{
		S3Object: s3object,
		Time:     time.Now(),
	}

	obj, err := dynamodbattribute.MarshalMap(objMap)

	if err != nil {
		return fmt.Errorf("Marshalling DynamoDB object failed: %s", err)
	}

	// add object to dynamodb using conditional
	// if the object exists, no write happens
	input := &dynamodb.PutItemInput{
		Item:                obj,
		TableName:           aws.String(d.TableName),
		ConditionExpression: aws.String("attribute_not_exists(S3Object)"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// we want this to happen if object already exists
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return fmt.Errorf("Item exists in Dynamo: %s", err)
			} else {
				return fmt.Errorf("PutItem failed: %s", err)
			}

		}
		// if it is the conditional check, we can just pop out and
		// ignore this object!
		return nil
	}

	return nil
}

func (d *DynamoDBStater) IsProcessed(object string) (bool, error) {
	r, err := d.svc.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"S3Object": {
				S: aws.String(object),
			},
		},
		TableName:      aws.String(d.TableName),
		ConsistentRead: aws.Bool(true),
	})

	if r.Item == nil {
		return false, nil
	}
	return true, err
}

// FileStater is an implementation for indicating processing state using the
// local filesystem for backing storage.
type FileStater struct {
	*sync.Mutex
	StateDir         string
	Service          string
	BackfillInterval time.Duration
}

func NewFileStater(stateDir, service string) *FileStater {
	return &FileStater{
		Mutex:    &sync.Mutex{},
		StateDir: stateDir,
		Service:  service,
	}
}

func (f *FileStater) stateFile() string {
	return filepath.Join(f.StateDir, fmt.Sprintf(stateFileFormat, f.Service))
}

func (f *FileStater) processedObjects() (map[string]time.Time, error) {
	objs := make(map[string]time.Time)

	if _, err := os.Stat(f.stateFile()); os.IsNotExist(err) {
		// make sure file exists first run
		if err := ioutil.WriteFile(f.stateFile(), []byte(`{}`), 0644); err != nil {
			return objs, fmt.Errorf("Error writing file: %s", err)
		}

		return objs, nil
	}

	data, err := ioutil.ReadFile(f.stateFile())
	if err != nil {
		return objs, fmt.Errorf("Error reading object cursor file: %s", err)
	}

	if err := json.Unmarshal(data, &objs); err != nil {
		return objs, fmt.Errorf("Unmarshalling state file JSON failed: %s", err)
	}

	return objs, nil
}

func (f *FileStater) ProcessedObjects() (map[string]time.Time, error) {
	f.Lock()
	defer f.Unlock()
	return f.processedObjects()
}

func (f *FileStater) IsProcessed(object string) (bool, error) {
	processedObjects, err := f.ProcessedObjects()
	if err != nil {
		return false, err
	}
	_, present := processedObjects[object]

	return present, nil
}

func (f *FileStater) SetProcessed(object string) error {
	f.Lock()
	defer f.Unlock()

	processedObjects, err := f.processedObjects()
	if err != nil {
		return err
	}

	processedObjects[object] = time.Now()

	processedData, err := json.Marshal(processedObjects)
	if err != nil {
		return fmt.Errorf("Marshalling JSON failed: %s", err)
	}

	if err := ioutil.WriteFile(f.stateFile(), processedData, 0644); err != nil {
		return fmt.Errorf("Writing file failed: %s", err)
	}

	return nil
}
