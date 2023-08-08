package blockchain

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/crypto"
)

type OperationType uint

const (
	UNKOWN_OPERATION OperationType = iota
	CREATE_FILE
	APPEND_RECORD
	DELETE_FILE
)

type OperationMsg struct {
	OpType    OperationType `json:"op_type"`
	OpFrom    Address       `json:"op_from"`
	Signature struct {
		Hash []byte `json:"hash"`
		Sig  []byte `json:"sig"`
	} `json:"sig"`
	// encoded operation
	Op   []byte `json:"op"`
	Hash string `json:"hash"`
}

type CreateFileOp struct {
	Fname string `json:"fname"`
}

type AppendRecordOp struct {
	Fname  string `json:"fname"`
	Record Record `json:"record"`
}

func ValidateOperation(op OperationMsg, address Address, accountStorage *AccountStorage) (bool, error) {
	// check if op is valid
	account, err := crypto.AccountFromAddress(string(address))
	if err != nil {
		return false, nil
	}

	switch op.OpType {
	case CREATE_FILE:
		// check balance
		if accountStorage.GetBalance() < uint64(NumCoinsPerFileCreate) {
			return false, ErrInsufficientFunds
		}
	case APPEND_RECORD:
		if accountStorage.GetBalance() < uint64(NumCoinsPerFileAppend) {
			return false, ErrInsufficientFunds
		}
	case DELETE_FILE:
		// check balance
		if accountStorage.GetBalance() < uint64(NumCoinsPerFileDelete) {
			return false, ErrInsufficientFunds
		}
	default:
		return false, nil
	}

	if op.OpFrom == "" {
		return false, nil
	}

	if op.OpFrom != address {
		return false, nil
	}

	// TODO: think about more stuff to validate, aka rfs operations

	return account.Verify(op.Signature.Hash, op.Signature.Sig)
}

func (op *OperationMsg) GenerateHash() {
	hash := md5.New()
	hash.Write(append(op.Op, op.Signature.Hash...))
	op.Hash = hex.EncodeToString(hash.Sum(nil))
}
