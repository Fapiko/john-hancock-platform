package contracts

import "errors"

type KeyAlgorithm int

const (
	ECDSA KeyAlgorithm = iota
	ED25519
	RSA
	Unknown
)

var KeyAlgorithmStrings = map[KeyAlgorithm]string{
	ECDSA:   "ECDSA",
	ED25519: "ED25519",
	RSA:     "RSA",
}

func (a KeyAlgorithm) String() string {
	return KeyAlgorithmStrings[a]
}

func AlgorithmFromString(algorithm string) (KeyAlgorithm, error) {
	switch algorithm {
	case "ECDSA":
		return ECDSA, nil
	case "ED25519":
		return ED25519, nil
	case "RSA":
		return RSA, nil
	default:
		return Unknown, errors.New("unknown algorithm")
	}
}
