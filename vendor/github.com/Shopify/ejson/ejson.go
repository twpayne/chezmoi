// Package ejson implements the primary interface to interact with ejson
// documents and keypairs. The CLI implemented by cmd/ejson is a fairly thin
// wrapper around this package.
package ejson

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Shopify/ejson/crypto"
	"github.com/Shopify/ejson/json"
)

// GenerateKeypair is used to create a new ejson keypair. It returns the keys as
// hex-encoded strings, suitable for printing to the screen. hex.DecodeString
// can be used to load the true representation if necessary.
func GenerateKeypair() (pub string, priv string, err error) {
	var kp crypto.Keypair
	if err := kp.Generate(); err != nil {
		return "", "", err
	}
	return kp.PublicString(), kp.PrivateString(), nil
}

// Encrypt reads all contents from 'in', extracts the pubkey
// and performs the requested encryption operation, writing
// the resulting data to 'out'.
// Returns the number of bytes written and any error that might have
// occurred.
func Encrypt(in io.Reader, out io.Writer) (int, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return -1, err
	}

	var myKP crypto.Keypair
	if err = myKP.Generate(); err != nil {
		return -1, err
	}

	data, err = json.CollapseMultilineStringLiterals(data)
	if err != nil {
		return -1, err
	}

	pubkey, err := json.ExtractPublicKey(data)
	if err != nil {
		return -1, err
	}

	encrypter := myKP.Encrypter(pubkey)
	walker := json.Walker{
		Action: encrypter.Encrypt,
	}

	newdata, err := walker.Walk(data)
	if err != nil {
		return -1, err
	}

	return out.Write(newdata)
}

// EncryptFileInPlace takes a path to a file on disk, which must be a valid EJSON file
// (see README.md for more on what constitutes a valid EJSON file). Any
// encryptable-but-unencrypted fields in the file will be encrypted using the
// public key embdded in the file, and the resulting text will be written over
// the file present on disk.
func EncryptFileInPlace(filePath string) (int, error) {
	var fileMode os.FileMode
	if stat, err := os.Stat(filePath); err == nil {
		fileMode = stat.Mode()
	} else {
		return -1, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return -1, err
	}

	var outBuffer bytes.Buffer

	written, err := Encrypt(file, &outBuffer)
	if err != nil {
		return -1, err
	}

	if err = file.Close(); err != nil {
		return -1, err
	}

	if err := ioutil.WriteFile(filePath, outBuffer.Bytes(), fileMode); err != nil {
		return -1, err
	}

	return written, nil
}

// Decrypt reads an ejson stream from 'in' and writes the decrypted data to 'out'.
// The private key is expected to be under 'keydir'.
// Returns error upon failure, or nil on success.
func Decrypt(in io.Reader, out io.Writer, keydir string, userSuppliedPrivateKey string) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	pubkey, err := json.ExtractPublicKey(data)
	if err != nil {
		return err
	}

	privkey, err := findPrivateKey(pubkey, keydir, userSuppliedPrivateKey)
	if err != nil {
		return err
	}

	myKP := crypto.Keypair{
		Public:  pubkey,
		Private: privkey,
	}

	decrypter := myKP.Decrypter()
	walker := json.Walker{
		Action: decrypter.Decrypt,
	}

	newdata, err := walker.Walk(data)
	if err != nil {
		return err
	}

	_, err = out.Write(newdata)

	return err
}

// DecryptFile takes a path to an encrypted EJSON file and returns the data
// decrypted. The public key used to encrypt the values is embedded in the
// referenced document, and the matching private key is searched for in keydir.
// There must exist a file in keydir whose name is the public key from the
// EJSON document, and whose contents are the corresponding private key. See
// README.md for more details on this.
func DecryptFile(filePath, keydir string, userSuppliedPrivateKey string) ([]byte, error) {
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var outBuffer bytes.Buffer

	err = Decrypt(file, &outBuffer, keydir, userSuppliedPrivateKey)

	return outBuffer.Bytes(), err
}

func readPrivateKeyFromDisk(pubkey [32]byte, keydir string) (privkey string, err error) {
	keyFile := fmt.Sprintf("%s/%x", keydir, pubkey)
	var fileContents []byte
	fileContents, err = ioutil.ReadFile(keyFile)
	if err != nil {
		err = fmt.Errorf("couldn't read key file (%s)", err.Error())
		return
	}
	privkey = string(fileContents)
	return
}

func findPrivateKey(pubkey [32]byte, keydir string, userSuppliedPrivateKey string) (privkey [32]byte, err error) {
	var privkeyString string
	if userSuppliedPrivateKey != "" {
		privkeyString = userSuppliedPrivateKey
	} else {
		privkeyString, err = readPrivateKeyFromDisk(pubkey, keydir)
		if err != nil {
			return privkey, err
		}
	}

	privkeyBytes, err := hex.DecodeString(strings.TrimSpace(privkeyString))
	if err != nil {
		return
	}

	if len(privkeyBytes) != 32 {
		err = fmt.Errorf("invalid private key")
		return
	}
	copy(privkey[:], privkeyBytes)
	return
}
