package random

import (
	"crypto/rand"

	"github.com/c2pc/config-migrate/replacer"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@()_+-=."
const letters2 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	replacer.Register("___random___", randomReplacer(16))
	replacer.Register("___random8___", randomReplacer(8))
	replacer.Register("___random32___", randomReplacer(32))
	replacer.Register("___random64___", randomReplacer(64))
}

func randomReplacer(n int) func() string {
	return func() string {
		bytes := make([]byte, n)
		_, err := rand.Read(bytes)
		if err != nil {
			return ""
		}

		for i, b := range bytes {
			if i == 0 || i == len(bytes)-1 {
				bytes[i] = letters2[b%byte(len(letters2))]
			} else {
				bytes[i] = letters[b%byte(len(letters))]
			}
		}

		return string(bytes)
	}
}
