package network

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
)

type Msg struct {
	Op         Operation `json:"op"`
	Data       []byte    `json:"data"`
	AddrFrom   string    `json:"addr_from"`
	ReqID      string    `json:"req_id"`
	ResponseID string    `json:"response_id"`
}

type FullBlockchainResponse struct {
	Chain []blockchain.Block `json:"chain"`
}

type BlockRequest struct {
	Number uint64 `json:"number"`
}

type BlockResponse struct {
	Block blockchain.Block `json:"block"`
}

type OperationResponse struct {
	Op blockchain.OperationMsg `json:"op"`
}

type MemPoolResponse struct {
	MemPool map[string]blockchain.OperationMsg `json:"mempool"`
}

type HeightResponse struct {
	Height uint64 `json:"height"`
}

type PeersResponse struct {
	Peers []string `json:"peers"`
}

func NewMsg(op Operation, data any, addrFrom string, respID ...string) (Msg, error) {
	var reqid string = generateRequestID(op, data, addrFrom)
	var respid string
	if len(respID) > 0 {
		respid = respID[0]
	}

	d, err := json.Marshal(data)
	if err != nil {
		return Msg{}, err
	}
	msg := Msg{
		Op:         op,
		Data:       d,
		AddrFrom:   addrFrom,
		ReqID:      reqid,
		ResponseID: respid,
	}

	return msg, nil
}

func DecodeMsg(b []byte) (Msg, error) {
	var msg Msg
	err := json.Unmarshal(b, &msg)
	return msg, err
}

func generateRequestID(op Operation, data any, addrFrom string) string {
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d%v%s%s", op, data, addrFrom, time.Now().String())))
	return hex.EncodeToString(hash.Sum(nil))
}

func (m *Msg) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}
