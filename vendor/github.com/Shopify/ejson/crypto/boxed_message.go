package crypto

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
)

var messageParser = regexp.MustCompile("\\AEJ\\[(\\d):([A-Za-z0-9+=/]{44}):([A-Za-z0-9+=/]{32}):(.+)\\]\\z")

// boxedMessage dumps and loads the wire format for encrypted messages. The
// schema is fairly simple:
//
//   "EJ["
//   SchemaVersion ( "1" )
//   ":"
//   EncrypterPublic :: base64-encoded 32-byte key
//   ":"
//   Nonce :: base64-encoded 24-byte nonce
//   ":"
//   Box :: base64-encoded encrypted message
//   "]"
type boxedMessage struct {
	SchemaVersion   int
	EncrypterPublic [32]byte
	Nonce           [24]byte
	Box             []byte
}

// IsBoxedMessage tests whether a value is formatted using the boxedMessage
// format. This can be used to determine whether a string value requires
// encryption or is already encrypted.
func IsBoxedMessage(data []byte) bool {
	return messageParser.Find(data) != nil
}

// Dump dumps to the wire format
func (b *boxedMessage) Dump() []byte {
	pub := base64.StdEncoding.EncodeToString(b.EncrypterPublic[:])
	nonce := base64.StdEncoding.EncodeToString(b.Nonce[:])
	box := base64.StdEncoding.EncodeToString(b.Box)

	str := fmt.Sprintf("EJ[%d:%s:%s:%s]",
		b.SchemaVersion, pub, nonce, box)
	return []byte(str)
}

// Load restores from the wire format.
func (b *boxedMessage) Load(from []byte) error {
	var ssver, spub, snonce, sbox string
	var err error

	allMatches := messageParser.FindAllStringSubmatch(string(from), -1) // -> [][][]byte
	if len(allMatches) != 1 {
		return fmt.Errorf("invalid message format")
	}
	matches := allMatches[0]
	if len(matches) != 5 {
		return fmt.Errorf("invalid message format")
	}

	ssver = matches[1]
	spub = matches[2]
	snonce = matches[3]
	sbox = matches[4]

	b.SchemaVersion, err = strconv.Atoi(ssver)
	if err != nil {
		return err
	}

	pub, err := base64.StdEncoding.DecodeString(spub)
	if err != nil {
		return err
	}
	pubBytes := []byte(pub)
	if len(pubBytes) != 32 {
		return fmt.Errorf("public key invalid")
	}
	var public [32]byte
	copy(public[:], pubBytes[0:32])
	b.EncrypterPublic = public

	nnc, err := base64.StdEncoding.DecodeString(snonce)
	if err != nil {
		return err
	}
	nonceBytes := []byte(nnc)
	if len(nonceBytes) != 24 {
		return fmt.Errorf("nonce invalid")
	}
	var nonce [24]byte
	copy(nonce[:], nonceBytes[0:24])
	b.Nonce = nonce

	box, err := base64.StdEncoding.DecodeString(sbox)
	if err != nil {
		return err
	}
	b.Box = []byte(box)

	return nil
}
