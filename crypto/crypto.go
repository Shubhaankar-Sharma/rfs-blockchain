package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"time"
)

type Account struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  [64]byte
	Address    string
	rwMutex    sync.RWMutex
}

func NewAccount() *Account {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	x := [32]byte(private.PublicKey.X.Bytes())
	y := [32]byte(private.PublicKey.Y.Bytes())

	pub := [64]byte(append(x[:], y[:]...))

	// I don't see a reason why we should hash the public key
	// If we use an account level nonce, then we can just use the public key as the address
	// hence, im not hashing, and just using the hex form for easy readability
	addy := hex.EncodeToString(pub[:])

	return &Account{
		PrivateKey: private,
		PublicKey:  pub,
		Address:    addy,
	}
}

func AccountFromAddress(address string) (*Account, error) {
	if len(address) != 128 {
		return nil, errors.New("invalid address")
	}

	pub, err := hex.DecodeString(address)
	if err != nil {
		panic(err)
	}

	return &Account{
		PublicKey: [64]byte(pub),
		Address:   address,
	}, nil
}

func (a *Account) Sign(msg []byte) (hash []byte, sig []byte, err error) {
	hash = append([]byte(time.Now().String()), []byte("SignedMessage:")...)
	hash = append(hash, msg...)
	sig, err = ecdsa.SignASN1(rand.Reader, a.PrivateKey, hash)
	return hash, sig, err
}

func (a *Account) Verify(hash []byte, sig []byte) (bool, error) {
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(a.PublicKey[:32]),
		Y:     new(big.Int).SetBytes(a.PublicKey[32:]),
	}

	verify := ecdsa.VerifyASN1(pubKey, hash, sig)
	if !verify {
		return false, errors.New("invalid signature")
	}

	return true, nil
}
