package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type TFEiface interface {
	Backup(context.Context, s3manageriface.UploaderAPI, *DataKey, *Config) error
}

type TFE struct {
	token string
	host  string
}

func NewTFE(token, host string) TFEiface {
	return &TFE{
		token: token,
		host:  host,
	}
}

func (t *TFE) createBackupRequest(ctx context.Context, password string) (*http.Request, error) {
	body, err := json.Marshal(map[string]string{"password": password})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal ")
	}

	url := fmt.Sprintf("%s/_backup/api/v1/backup", t.host)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)

	// Add authentication token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.token))

	return req, errors.Wrap(err, "could not create backup request")
}

func (t *TFE) Backup(
	ctx context.Context,
	s3 s3manageriface.UploaderAPI,
	dataKey *DataKey,
	config *Config,
) error {

	// ask for backup
	req, err := t.createBackupRequest(ctx, dataKey.Plaintext)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not perform backup request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received non 200 status code %d", resp.StatusCode)
	}

	// calculate s3 object key
	key := path.Join(config.S3Prefix, time.Now().Format("2006/01/02"), uuid.NewString())
	// tag backup with ciphertext
	// to fetch plaintext, caller must decrypt through KMS
	tags := url.Values{}
	tags.Add("base64_ciphertext", dataKey.Ciphertext)

	// report how many bytes read from resp
	body := &Report{reader: resp.Body}

	// streaming upload to S3
	_, err = s3.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:  &config.S3Bucket,
		Key:     &key,
		Tagging: aws.String(tags.Encode()),
		Body:    body,
	})

	return errors.Wrap(err, "could not upload backup")
}
