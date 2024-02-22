package kmsauth

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthContextValidate(t *testing.T) {
	a := assert.New(t)
	ac := &AuthContextV1{From: "foo"}
	a.NotNil(ac.Validate())
	ac = &AuthContextV1{From: "foo", To: "bar"}
	a.Nil(ac.Validate())
	ac2 := &AuthContextV2{From: "foo"}
	a.NotNil(ac2.Validate())
	ac2 = &AuthContextV2{From: "foo", To: "bar"}
	a.NotNil(ac2.Validate())
	ac2 = &AuthContextV2{From: "foo", To: "bar", UserType: "foobar"}
	a.Nil(ac2.Validate())

	ac = nil
	a.NotNil(ac.Validate())
	ac2 = nil
	a.NotNil(ac2.Validate())
}

func TestAuthContextGetUsername(t *testing.T) {
	a := assert.New(t)
	ac := AuthContextV1{To: "foo"}
	a.Equal(ac.GetUsername(), "")
	ac = AuthContextV1{From: "foo", To: "bar"}
	a.Equal(ac.GetUsername(), "foo")
	ac2 := AuthContextV2{To: "foo"}
	a.Equal(ac2.GetUsername(), "2//")
	ac2 = AuthContextV2{From: "foo", To: "bar"}
	a.Equal(ac2.GetUsername(), "2//foo")
	ac2 = AuthContextV2{From: "foo", To: "bar", UserType: "gas"}
	a.Equal(ac2.GetUsername(), "2/gas/foo")
}

func TestAuthContextGetKSMContext(t *testing.T) {
	a := assert.New(t)

	foo := "foo"
	bar := "bar"
	baz := "baz"

	ac := AuthContextV1{From: foo, To: bar}
	expected := map[string]*string{
		"from": &foo,
		"to":   &bar,
	}
	a.True(reflect.DeepEqual(ac.GetKMSContext(), expected))

	ac2 := AuthContextV2{From: foo, To: bar, UserType: baz}
	expected = map[string]*string{
		"from":      &foo,
		"to":        &bar,
		"user_type": &baz,
	}
	a.True(reflect.DeepEqual(ac2.GetKMSContext(), expected))
}

func TestTokenTimeMarshal(t *testing.T) {
	a := assert.New(t)

	tiempo := time.Time{}
	tiempo = tiempo.Add(1 * time.Minute)

	tc := TokenCache{
		Token: Token{
			NotBefore: TokenTime{Time: tiempo},
		},
	}

	b, err := json.Marshal(tc)
	a.Nil(err)
	a.Equal(string(b), "{\"token\":{\"not_before\":\"0001-01-01T00:01:00Z\",\"not_after\":\"0001-01-01T00:00:00Z\"}}")
}

func TestNewToken(t *testing.T) {
	a := assert.New(t)

	// Correctly accounts for skew
	token := NewToken(0 * time.Minute)
	a.NotNil(token)
	a.Equal(token.NotAfter, token.NotBefore)

	// Goes to the future
	token = NewToken(100 * time.Minute)
	a.NotNil(token)
	a.True(token.NotAfter.After(time.Now().UTC()))

}
