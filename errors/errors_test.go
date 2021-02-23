package errors

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPublicErrors(t *testing.T) {
	r := require.New(t)

	opaqueError := func(base string, public string) error {
		baseError := fmt.Errorf(base)
		e := PublicWrap(baseError, public)
		return e
	}

	err := opaqueError("base error", "public message")

	publicMsg := GetPublicMessage(err)
	r.NotNil(publicMsg)
	r.Equal("public message", *publicMsg)

	baseErr := GetInternalError(err)
	r.NotNil(baseErr)
	r.Equal("base error", baseErr.Error())
	r.Equal("base error", errors.Cause(baseErr).Error())

	// not a publicError
	err = fmt.Errorf("scary error")
	publicMsg = GetPublicMessage(err)
	r.Nil(publicMsg)
}

func TestPublicWrapNilError(t *testing.T) {
	r := require.New(t)
	r.NoError(PublicWrap(nil, "error that should be nill"))
}

func TestPublicWrapfNilError(t *testing.T) {
	r := require.New(t)
	r.NoError(PublicWrapf(nil, "error that should be nill"))
}

func TestPublicWrapf(t *testing.T) {
	r := require.New(t)
	err := PublicWrapf(
		fmt.Errorf("blabla"),
		"%s",
		"foobar",
	)
	r.Error(err)

	r.Equal("foobar", *GetPublicMessage(err))
	r.Equal("foobar", err.Error())
}

func TestGetInternalErrorNilError(t *testing.T) {
	r := require.New(t)
	var nilError *publicError

	r.NoError(nilError.GetInternalError())
}
