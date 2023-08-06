package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
)

type Account struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  [64]byte
	Address    string
	rwMutex    sync.RWMutex
	Nonce      uint64
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
		Nonce:      0,
	}
}

func AccountFromAddress(address string, nonce uint64) (*Account, error) {
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
		Nonce:     nonce,
	}, nil
}

func (a *Account) IncrementNonce() {
	a.rwMutex.Lock()
	defer a.rwMutex.Unlock()
	a.Nonce += 1
}

func (a *Account) GetNonce() uint64 {
	a.rwMutex.RLock()
	defer a.rwMutex.RUnlock()
	return a.Nonce
}

func (a *Account) Sign(msg []byte) (hash []byte, sig []byte, err error) {
	nonceBytes := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(nonceBytes, a.GetNonce())
	hash = append(nonceBytes, []byte("SignedMessage:")...)
	hash = append(hash, msg...)
	sig, err = ecdsa.SignASN1(rand.Reader, a.PrivateKey, hash)
	return hash, sig, err
}

func (a *Account) Verify(hash []byte, sig []byte, correctNonce uint64) bool {
	nonceBytes := hash[:binary.MaxVarintLen64]
	// bytes to uint64
	nonce, _ := binary.Uvarint(nonceBytes)
	if nonce != correctNonce {
		return false
	}

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(a.PublicKey[:32]),
		Y:     new(big.Int).SetBytes(a.PublicKey[32:]),
	}

	return ecdsa.VerifyASN1(pubKey, hash, sig)
}
