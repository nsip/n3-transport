// crypto.go

package n3crypto

import (
	"crypto/rand"

	"github.com/kevinburke/nacl"
	"github.com/kevinburke/nacl/box"
	"github.com/nsip/n3-transport/messages"
	"github.com/nsip/n3-transport/messages/pb"
	"github.com/shengdoushi/base58"
)

// select an encoding alphabet compatible with
// other services in the n3 ecosystem
var encodingAlphabet = base58.IPFSAlphabet

//
// package uses go-nacl encryption so that tuples are
// encrypted and signed in one operation
//

//
// returns nacl box public & private keys
//
func GenerateKeys() (b58pubkey string, b58privkey string, err error) {

	publickey, privatekey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	b58pubkey = base58.Encode(publickey[:], encodingAlphabet)
	b58privkey = base58.Encode(privatekey[:], encodingAlphabet)

	return b58pubkey, b58privkey, nil
}

//
// decodes a b58 string of the kay back into a nacl key
//
func B58Decode(key string) (nacl.Key, error) {

	byteKey, err := base58.Decode(key, encodingAlphabet)
	if err != nil {
		return nil, err
	}
	newKey := nacl.NewKey()
	copy(newKey[:], byteKey)

	return newKey, nil
}

//
// takes a tuple and encrypts with nacl.box
//
func EncryptTuple(t *pb.SPOTuple, recipientPublicKey, senderPrivateKey string) ([]byte, error) {

	// decode keys
	recipientKey, err := B58Decode(recipientPublicKey)
	senderKey, err := B58Decode(senderPrivateKey)
	if err != nil {
		return nil, err
	}

	// encode tuple to bytes using protobuf
	tupleBytes, err := messages.EncodeTuple(t)
	if err != nil {
		return nil, err
	}

	// encrypt the tuple bytes
	encryptedTuple := box.EasySeal(tupleBytes, recipientKey, senderKey)

	return encryptedTuple, nil
}

//
// Decrypts a data tuple received as part of an N3 message
//
func DecryptTuple(tbytes []byte, senderPublicKey, recipientPrivateKey string) (*pb.SPOTuple, error) {

	// decode keys
	senderKey, err := B58Decode(senderPublicKey)
	recipientKey, err := B58Decode(recipientPrivateKey)
	if err != nil {
		return nil, err
	}

	// decrypt the tuple bytes
	unencrpytedTupleBytes, err := box.EasyOpen(tbytes, senderKey, recipientKey)
	if err != nil {
		return nil, err
	}

	// decode the bytes from protobuf
	unencryptedTuple, err := messages.DecodeTuple(unencrpytedTupleBytes)
	if err != nil {
		return nil, err
	}

	return unencryptedTuple, nil

}
