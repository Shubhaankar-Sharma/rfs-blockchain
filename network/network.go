package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shubhaankar-Sharma/rfs-blockchain/blockchain"
)

type Network struct {
	ctx     context.Context
	cancel  context.CancelFunc
	running *int32

	addr string

	mu         sync.RWMutex
	peers      map[Addr]struct{}
	blockchain IBlockchain

	reqIDsListened map[string]struct{}

	registerBlockChan       chan<- blockchain.Block
	distributeBlockChan     <-chan blockchain.Block
	registerOperationChan   chan<- blockchain.OperationMsg
	distributeOperationChan <-chan blockchain.OperationMsg

	reqIDs map[string]chan Msg
}

func NewNetwork(listenAddr string, blockchain IBlockchain, peers ...Addr) *Network {
	running := int32(0)
	ctx, cancel := context.WithCancel(context.Background())
	peerMap := make(map[Addr]struct{})
	for _, peer := range peers {
		peerMap[peer] = struct{}{}
	}

	return &Network{
		ctx:            ctx,
		cancel:         cancel,
		blockchain:     blockchain,
		peers:          peerMap,
		running:        &running,
		reqIDs:         make(map[string]chan Msg),
		addr:           listenAddr,
		reqIDsListened: make(map[string]struct{}),
	}
}

func (n *Network) RegisterChannels() {
	if n.blockchain == nil {
		panic("blockchain cannot be nil")
	}

	n.registerBlockChan, n.distributeBlockChan = n.blockchain.GetBlockPubSubChans()
	n.registerOperationChan, n.distributeOperationChan = n.blockchain.GetOperationPubSubChans()
}

func (n *Network) Run() error {
	if n.IsRunning() {
		return errors.New("network already running")
	}

	if n.blockchain == nil {
		return errors.New("blockchain cannot be nil")
	}

	if n.registerBlockChan == nil || n.distributeBlockChan == nil || n.registerOperationChan == nil || n.distributeOperationChan == nil {
		return errors.New("blockchain not registered")
	}

	atomic.StoreInt32(n.running, 1)
	n.run()
	return nil
}

func (n *Network) run() {
	ln, err := net.Listen("tcp", n.addr)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go n.handleRequest(conn)
		}
	}
}

func (n *Network) IsRunning() bool {
	return atomic.LoadInt32(n.running) == 1
}

func (n *Network) Stop() {
	if !n.IsRunning() {
		return
	}
	n.cancel()
	atomic.StoreInt32(n.running, 0)
}

func (n *Network) handleRequest(conn net.Conn) {
	req, err := io.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}

	go n.reqHandler(req)
}

func (n *Network) reqHandler(req []byte) {
	msg, err := DecodeMsg(req)
	if err != nil {
		log.Println("error decoding request:", err)
		return
	}
	if msg.Op == UNKOWN {
		log.Println("unknown request")
		return
	}
	if msg.ReqID == "" {
		log.Println("request missing reqID")
		return
	}
	if msg.AddrFrom == "" {
		log.Println("request missing addrFrom")
		return
	}

	if msg.AddrFrom == n.addr {
		log.Println("request from self")
		return
	}

	n.mu.RLock()
	_, ok := n.reqIDsListened[msg.ReqID]
	n.mu.RUnlock()
	if ok {
		log.Println("request already listened to")
		return
	}

	n.mu.Lock()
	if _, ok := n.peers[Addr(msg.AddrFrom)]; !ok {
		n.peers[Addr(msg.AddrFrom)] = struct{}{}
	}
	n.reqIDsListened[msg.ReqID] = struct{}{}
	n.mu.Unlock()

	var isResponse bool
	if msg.ResponseID != "" {
		isResponse = true
	}

	// fmt.Println("request:", msg)

	if isResponse {
		n.mu.RLock()
		msgChan, ok := n.reqIDs[msg.ResponseID]
		n.mu.RUnlock()
		if !ok {
			log.Println("response for unknown request")
			return
		}
		msgChan <- msg
		return
	}

	switch msg.Op {
	// handle blocking responses that we could recieve
	case SEND_FULL_BLOCKCHAIN, SEND_MEM_POOL, SEND_HEIGHT:
		// ignore these requests cause they should ideally be responses
		return
	case SEND_PEERS:

	case SEND_BLOCK:

	case SEND_OPERATION:

	case GET_FULL_BLOCKCHAIN:
	case GET_BLOCK:
	case GET_BEST_HEIGHT:
		n.handleGetBestHeightRequest(&msg)
		// do it rn
	}
	// broadcast msg
	n.broadcastMsg(&msg)
}

func (n *Network) broadcastMsg(msg *Msg) {
	for peer := range n.peers {
		if peer == Addr(msg.AddrFrom) {
			continue
		} else if string(peer) == n.addr {
			continue
		}

		n.sendMsg(string(peer), msg)
	}
}

func (n *Network) handleGetBestHeightRequest(msg *Msg) {
	// send response
	if msg.Op != GET_BEST_HEIGHT {
		return
	}

	bestHeight := n.blockchain.CurrentHeight()
	resp := HeightResponse{
		Height: bestHeight,
	}
	respMsg, err := NewMsg(SEND_HEIGHT, resp, n.addr, msg.ReqID)
	if err != nil {
		log.Println("error creating response:", err)
		return
	}
	// fmt.Printf("--------------SENDING MESSAGE-----------------: %v %v To: %s\n", respMsg, resp, msg.AddrFrom)
	n.sendMsg(msg.AddrFrom, &respMsg)
}

func (n *Network) sendMsg(addr string, msg *Msg) error {
	data, err := msg.ToBytes()

	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) makeRequestWithResponse(addr string, msg Msg, waitFor time.Duration, foundResponses func([]Msg) (Msg, bool)) (*Msg, error) {
	data, err := msg.ToBytes()

	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		n.mu.Lock()
		delete(n.peers, Addr(addr))
		n.mu.Unlock()
		return nil, err
	}
	// dispatch listeners
	n.mu.Lock()
	reqID := msg.ReqID
	msgChan := make(chan Msg, 256)
	n.reqIDs[reqID] = msgChan
	n.reqIDsListened[reqID] = struct{}{}
	n.mu.Unlock()

	finalAnswer := make(chan Msg, 1)
	go func() {
		var msgs []Msg
		var final Msg
		defer func() {
			n.mu.Lock()
			delete(n.reqIDs, reqID)
			close(msgChan)
			n.mu.Unlock()
			finalAnswer <- final
		}()

		ctx, _ := context.WithDeadline(n.ctx, time.Now().Add(waitFor))
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgChan:
				if msg.ResponseID == reqID {
					// fmt.Println("got response", msg)
					msgs = append(msgs, msg)
					var ok bool
					final, ok = foundResponses(msgs)
					if ok {
						return
					}
				}
			}
		}
	}()

	_, err = io.Copy(conn, bytes.NewReader(data))
	conn.Close()
	if err != nil {
		return nil, err
	}
	response := <-finalAnswer
	if response.ResponseID == "" {
		return nil, errors.New("no response")
	}

	return &response, nil
}

func (n *Network) GetBestHeight() (uint64, string) {
	var mu sync.Mutex
	var bestHeights map[string]uint64 = make(map[string]uint64)
	var wg sync.WaitGroup
	for addr := range n.peers {
		ad := addr
		wg.Add(1)
		go func(addr Addr) {
			defer wg.Done()
			bestAddy, height, err := n.getBestHeight(addr)
			if err != nil {
				log.Println(err)
				return
			}
			mu.Lock()
			bestHeights[bestAddy] = height
			mu.Unlock()
		}(ad)
	}
	wg.Wait()

	var bestHeight uint64 = 0
	var bestHeightPeer string
	for addr, height := range bestHeights {
		if height > bestHeight {
			bestHeight = height
			bestHeightPeer = addr
		}
	}

	return bestHeight, bestHeightPeer
}

func (n *Network) getBestHeight(addr Addr) (string, uint64, error) {
	req, _ := NewMsg(GET_BEST_HEIGHT, nil, n.addr)

	foundResponses := func(msgs []Msg) (Msg, bool) {
		var best Msg
		var bestHeight uint64 = 0
		for _, msg := range msgs {
			// decode message to Height msg
			if msg.Op != SEND_HEIGHT {
				log.Println("invalid response")
				continue
			}

			var heightResponse HeightResponse
			err := json.Unmarshal(msg.Data, &heightResponse)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Println("response", heightResponse.Height, msg.AddrFrom)
			if heightResponse.Height > bestHeight {
				bestHeight = heightResponse.Height
				best = msg
			}
		}
		return best, false
	}

	resp, err := n.makeRequestWithResponse(string(addr), req, time.Second*5, foundResponses)

	if err != nil {
		log.Println("ERROR", err)
		return "", 0, err
	}

	if resp.Op != SEND_HEIGHT {
		log.Println("invalid response")
		return "", 0, errors.New("invalid response")
	}

	var heightResponse HeightResponse
	err = json.Unmarshal(resp.Data, &heightResponse)

	if err != nil {
		log.Println(err)
		return "", 0, err
	}

	return resp.AddrFrom, heightResponse.Height, nil
}

// func (n *Network) getFullBlockchain(addr Addr) ([]blockchain.Block, error) {
// 	req, _ := NewMsg(GET_FULL_BLOCKCHAIN, nil, n.addr)

// 	resp, err := n.makeRequest(string(addr), req)

// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	var msg Msg
// 	err = json.Unmarshal(resp, &msg)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	if msg.Op != SEND_FULL_BLOCKCHAIN {
// 		log.Println("invalid response")
// 		return nil, errors.New("invalid response")
// 	}

// 	var fullBlockchainResponse FullBlockchainResponse
// 	err = json.Unmarshal(msg.Data, &fullBlockchainResponse)

// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	return fullBlockchainResponse.Chain, nil
// }

// func (n *Network) getBlock()
