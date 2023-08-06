package blockchain

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
	Op []byte `json:"op"`
}

type CreateFileOp struct {
	Fname string `json:"fname"`
}

type AppendRecordOp struct {
	Fname  string `json:"fname"`
	Record Record `json:"record"`
}

func ValidateOperation(op OperationMsg) bool {
	// check if op is valid
	switch op.OpType {
	case CREATE_FILE:
		return true
	case APPEND_RECORD:
		return true
	case DELETE_FILE:
		return true
	default:
		return false
	}

	// account := GetAccount(op.OpFrom)
}
