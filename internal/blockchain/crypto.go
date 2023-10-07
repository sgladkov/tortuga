package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

const CompactPublicKeyLength = 65
const AddressBytesOffset = 12

func GeneratePrivateKey() ([]byte, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return crypto.FromECDSA(privateKey), nil
}

func PublicKeyFromPrivateKey(privKeyBytes []byte) ([]byte, error) {
	privateKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	return crypto.FromECDSAPub(publicKeyECDSA), nil
}

func AddressFromPublicKey(pubKey []byte) (string, error) {
	if len(pubKey) != CompactPublicKeyLength {
		return "", errors.New("invalid public key length")
	}
	hash := sha3.NewLegacyKeccak256()
	_, err := hash.Write(pubKey[1:]) // remove EC prefix
	if err != nil {
		return "", err
	}
	buf := hash.Sum(nil)
	address := fmt.Sprintf("0x%s", hex.EncodeToString(buf[AddressBytesOffset:]))

	return address, nil
}

func SignData(data []byte, privKeyBytes []byte) ([]byte, error) {
	privateKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, err
	}
	hash := sha3.NewLegacyKeccak256()
	_, err = hash.Write(data)
	if err != nil {
		return nil, err
	}
	buf := hash.Sum(nil)
	return crypto.Sign(buf, privateKey)
}

func RestoreAddressFromSignature(data []byte, signature []byte) (string, error) {
	hash := sha3.NewLegacyKeccak256()
	_, err := hash.Write(data)
	if err != nil {
		return "", err
	}
	buf := hash.Sum(nil)
	publicKey, err := crypto.Ecrecover(buf, signature)
	if err != nil {
		return "", err
	}
	return AddressFromPublicKey(publicKey)
}
