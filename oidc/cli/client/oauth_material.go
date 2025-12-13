package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type oauthMaterial struct {
	Nonce      string
	NonceBytes []byte

	State      string
	StateBytes []byte

	CodeVerifier  string
	CodeChallenge string
}

func newOauthMaterial() (*oauthMaterial, error) {
	generateURLSafeRandom := func(numBytes int) (string, error) {
		b := make([]byte, numBytes)
		_, err := rand.Read(b)
		if err != nil {
			return "", fmt.Errorf("could not read random bytes: %w", err)
		}
		return pkceBase64URLEncode(b), nil
	}

	nonce, err := generateURLSafeRandom(32)
	if err != nil {
		return nil, err
	}
	state, err := generateURLSafeRandom(32)
	if err != nil {
		return nil, err
	}
	codeVerifier, err := generateURLSafeRandom(64)
	if err != nil {
		return nil, err
	}

	codeChallengeBytes := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := pkceBase64URLEncode(codeChallengeBytes[:])

	return &oauthMaterial{
		Nonce:      nonce,
		NonceBytes: []byte(nonce),

		State:      state,
		StateBytes: []byte(state),

		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
	}, nil
}

func pkceBase64URLEncode(b []byte) string {
	r := base64.URLEncoding.EncodeToString(b)
	// https://auth0.com/docs/api-auth/tutorials/authorization-code-grant-pkce
	// For some reason have to replace these chars, we lose some entropy but that's ok
	r = strings.ReplaceAll(r, "+", "-")
	r = strings.ReplaceAll(r, "/", "_")
	r = strings.ReplaceAll(r, "=", "")
	return r
}
