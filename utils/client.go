package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

const (
	TronMainNet = "grpc.trongrid.io:50051"
	TronTestNet = "grpc.shasta.trongrid.io:50051"
)

type TronClient struct {
	Client     *client.GrpcClient
	privateKey *ecdsa.PrivateKey
	Address    string
}

func NewTronClient(nodeURL string) (*TronClient, error) {
	conn := client.NewGrpcClient(nodeURL)
	err := conn.Start(grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %v", err)
	}

	privateKey, err := LoadPrivateKeyFromEnv()
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	tronAddress := address.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).String()
	log.Println("eth address", address.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).Hex())

	return &TronClient{
		Client:     conn,
		privateKey: privateKey,
		Address:    tronAddress,
	}, nil
}

func (tc *TronClient) GetBalance() (*big.Int, error) {
	account, err := tc.Client.GetAccount(tc.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account information: %v", err)
	}

	return big.NewInt(account.Balance), nil
}

func (tc *TronClient) GetRefBlockInfo() (*RefBlockInfo, error) {
	block, err := tc.Client.GetNowBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	refBlockBytes := make([]byte, 2)
	blockNumberBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockNumberBytes, uint64(block.GetBlockHeader().GetRawData().GetNumber()))
	copy(refBlockBytes, blockNumberBytes[6:8])

	refBlockHash := make([]byte, 8)
	copy(refBlockHash, block.GetBlockid()[8:16])

	return &RefBlockInfo{
		RefBlockBytes: refBlockBytes,
		RefBlockHash:  refBlockHash,
	}, nil
}

func (tc *TronClient) CreateTransferOffline(toAddress string, amount *big.Int, refBlock *RefBlockInfo) (string, error) {
	tx, err := CreateTransferTransaction(tc.Address, toAddress, amount.Int64(), refBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create transfer transaction: %v", err)
	}

	signedTx, err := SignTransactions(tx, tc.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	result, err := tc.Client.Broadcast(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("broadcast failed: %s", result.Message)
	}

	txHash, err := GetTransactionHash(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction hash: %v", err)
	}

	return txHash, nil
}

func (tc *TronClient) CreateTriggerContractOffline(
	contractAddress, method, jsonString string,
	callValue, feeLimit int64,
	refBlock *RefBlockInfo,
) (string, error) {
	tx, err := CreateTriggerContractTransaction(tc.Address, contractAddress, method, jsonString, callValue, feeLimit, refBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create trigger contract transaction: %v", err)
	}

	resp, err := tc.Client.EstimateEnergy(
		tc.Address,
		contractAddress,
		method,
		jsonString,
		0,
		"",
		0,
	)
	if err != nil {
		return "", fmt.Errorf("failed to estimate energy: %v", err)
	}
	if !resp.Result.Result || resp.Result.Code != 0 {
		return "", fmt.Errorf("energy estimation failed: %s", resp.Result.Message)
	}

	signedTx, err := SignTransactions(tx, tc.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	rawData, err := proto.Marshal(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal raw data: %v", err)
	}
	fmt.Println("rawData", hex.EncodeToString(rawData))

	req := struct {
		Transaction string `json:"transaction"`
	}{hex.EncodeToString(rawData)}

	postBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", "https://api.shasta.trongrid.io/wallet/broadcasthex", bytes.NewReader(postBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Content-Type", "application/json")

	dialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: 2000,
		MaxConnsPerHost:     2000,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	httpclient := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}
	httpResp, err := httpclient.Do(httpReq)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println("response body", string(body))

	// result, err := tc.Client.Broadcast(signedTx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to broadcast transaction: %v", err)
	// }

	// if result.Code != 0 {
	// 	return "", fmt.Errorf("broadcast failed: %s", result.Message)
	// }

	txHash, err := GetTransactionHash(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction hash: %v", err)
	}

	return txHash, nil
}

func (tc *TronClient) SendTx(tx *api.TransactionExtention) (string, error) {
	signedTx, err := SignTransactions(tx.Transaction, tc.privateKey)
	if err != nil {
		return "", err
	}

	result, err := tc.Client.Broadcast(signedTx)
	if err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("broadcast failed: %s", result.Message)
	}

	return hex.EncodeToString(tx.Txid), nil
}
