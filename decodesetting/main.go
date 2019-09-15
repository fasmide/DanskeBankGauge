package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// MagicLocation is where we start to find bytes from the key material
// notice this seems to change from time to time
// const MagicLocation = 0x108
const MagicLocation = 0x1DD

// this app is about loading some key material and then decrypt strings read from stdin
func main() {
	if len(os.Args) == 1 {
		log.Fatal("please add argument to key file")
	}

	keyPath := os.Args[1]
	log.Printf("Using %s", keyPath)

	key, err := NewKey(keyPath)
	if err != nil {
		log.Fatalf("could not load key: %s", err)
	}

	// bufio NewScanner defaults to newline
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		decoded, err := base64.StdEncoding.DecodeString(scanner.Text())
		if err != nil {
			log.Fatalf("unable to base64 decode input: %s", err)
		}

		decrypted, err := key.Decrypt(decoded)
		if err != nil {
			log.Fatalf("unable to decrypt: %s", err)
		}

		fmt.Println(decrypted)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("reading standard input: %s", err)
	}
}

// Key holds decryption key and allows decryption of strings
type Key struct {
	key []byte
	pos int
}

// NewKey returns a new initialized key
func NewKey(path string) (*Key, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %s", path, err)
	}

	return &Key{key: data, pos: MagicLocation}, nil
}

func (k *Key) decryptionByte(i int) (byte, error) {
	k.pos = k.pos + i
	if i > 0 {
		k.pos++
	}

	if len(k.key) < k.pos {
		return 0, fmt.Errorf("this key is not large enough to decode pos %d", k.pos)
	}

	return k.key[k.pos], nil
}

// Decrypt decrypts input
func (k *Key) Decrypt(input []byte) (string, error) {
	var result string
	for i, char := range input {
		b, err := k.decryptionByte(i)
		if err != nil {
			return "", err
		}

		result += string(char ^ b)
	}

	// reset position
	k.pos = MagicLocation

	return result, nil
}
