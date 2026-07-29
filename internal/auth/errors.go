package auth

import "errors"

var (
	ErrInvalidNonce     = errors.New("invalid or expired nonce")
	ErrInvalidSignature = errors.New("signature verification failed")
	ErrInvalidSession   = errors.New("invalid or expired session")
)
