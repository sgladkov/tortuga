package blockchain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrypto_SignVerify(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	require.NoError(t, err)
	pubKey, err := PublicKeyFromPrivateKey(privKey)
	require.NoError(t, err)
	address, err := AddressFromPublicKey(pubKey)
	require.NoError(t, err)
	data := []byte("some test data")
	signature, err := SignData(data, privKey)
	require.NoError(t, err)
	restoredAddress, err := RestoreAddressFromSignature(data, signature)
	require.NoError(t, err)
	require.Equal(t, address, restoredAddress)

	privKey1, err := GeneratePrivateKey()
	require.NoError(t, err)
	signature1, err := SignData(data, privKey1)
	require.NoError(t, err)
	restoredAddress1, err := RestoreAddressFromSignature(data, signature1)
	require.NoError(t, err)
	require.NotEqual(t, address, restoredAddress1)

	restoredAddress2, err := RestoreAddressFromSignature([]byte("other test data"), signature)
	require.NoError(t, err)
	require.NotEqual(t, address, restoredAddress2)
}
