package errors

import (
	"fmt"
	"testing"

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

func TestChainOfErrors(t *testing.T) {
	r := require.New(t)

	pubErr1 := PublicWrap(fmt.Errorf("error1"), "public message1")
	err2 := fmt.Errorf("error2")
	err3 := fmt.Errorf("error3")
	pubErr2 := PublicWrap(fmt.Errorf("error4"), "public message2")
	err5 := fmt.Errorf("error5")

	//err5 -> pubErr2 -> err3 -> err2 -> pubErr1
	err := fmt.Errorf("%w: %w: %w: %w: %w", err5, pubErr2, err3, err2, pubErr1)
	msg := GetPublicMessage(err)
	r.NotNil(msg)
	// gets the first public message off the chain
	r.Equal("public message2", *msg)

	// gets the first internal error off the chain
	intErr := GetInternalError(err)
	r.NotNil(intErr)
	r.Equal("error4", intErr.Error())
}

func TestIsPublicError(t *testing.T) {
	r := require.New(t)

	pubErr := PublicWrap(fmt.Errorf("error"), "public message")

	// any error wrapping a publicError is a public error
	r.True(IsPublicError(pubErr))
	r.True(IsPublicError(fmt.Errorf("error: %w", pubErr)))
	r.True(IsPublicError(fmt.Errorf("error: %w", fmt.Errorf("error: %w", pubErr))))

	// errors without a publicError in the chain are not public errors
	r.False(IsPublicError(nil))
	r.False(IsPublicError(fmt.Errorf("error")))
	r.False(IsPublicError(fmt.Errorf("error: %w", fmt.Errorf("error"))))
}
