package runner

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
)

type DataKey struct {
	Plaintext  string
	Ciphertext string
}

// Generates a Data key from an AWS KMS key we can use to encrypt our backups
func GenerateDataKey(
	ctx context.Context,
	k kmsiface.KMSAPI,
	kmsKeyARN string,
) (*DataKey, error) {
	output, err := k.GenerateDataKeyWithContext(ctx, &kms.GenerateDataKeyInput{
		KeyId:   aws.String(kmsKeyARN),
		KeySpec: aws.String("AES_256"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not kms generate data key")
	}

	// NOTE the base64 encoding!!
	return &DataKey{
		Plaintext:  base64.StdEncoding.EncodeToString(output.Plaintext),
		Ciphertext: base64.StdEncoding.EncodeToString(output.CiphertextBlob),
	}, nil
}
