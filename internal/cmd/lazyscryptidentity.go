package cmd

import (
	"errors"
	"fmt"

	"filippo.io/age"
)

// This is copied from https://github.com/FiloSottile/age/blob/6ad4560f4afc3fe46b6cda0bc568e50b89a22e4c/cmd/age/encrypted_keys.go#L15-L51

// LazyScryptIdentity is an age.Identity that requests a passphrase only if it
// encounters an scrypt stanza. After obtaining a passphrase, it delegates to
// ScryptIdentity.
type LazyScryptIdentity struct {
	Passphrase func() (string, error)
}

var _ age.Identity = &LazyScryptIdentity{}

func (i *LazyScryptIdentity) Unwrap(stanzas []*age.Stanza) (fileKey []byte, err error) {
	for _, s := range stanzas {
		if s.Type == "scrypt" && len(stanzas) != 1 {
			return nil, errors.New("an scrypt recipient must be the only one")
		}
	}
	if len(stanzas) != 1 || stanzas[0].Type != "scrypt" {
		return nil, age.ErrIncorrectIdentity
	}
	pass, err := i.Passphrase()
	if err != nil {
		return nil, fmt.Errorf("could not read passphrase: %w", err)
	}
	ii, err := age.NewScryptIdentity(pass)
	if err != nil {
		return nil, err
	}
	fileKey, err = ii.Unwrap(stanzas)
	if errors.Is(err, age.ErrIncorrectIdentity) {
		// ScryptIdentity returns ErrIncorrectIdentity for an incorrect
		// passphrase, which would lead Decrypt to returning "no identity
		// matched any recipient". That makes sense in the API, where there
		// might be multiple configured ScryptIdentity. Since in cmd/age there
		// can be only one, return a better error message.
		return nil, errors.New("incorrect passphrase")
	}
	return fileKey, err
}
