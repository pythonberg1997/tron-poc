package utils

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/stretchr/testify/require"
)

func TestCreateTx(t *testing.T) {
	nodeURL := TronTestNet
	tronClient, err := NewTronClient(nodeURL)
	if err != nil {
		log.Fatalf("failed to initialize Tron client: %v", err)
	}
	defer tronClient.Client.Stop()

	toAddress := "TGj1Ej1qRzL9feLTLhjwgxXF4Ct6GTWg2U"
	transferAmount := big.NewInt(100000)

	tx, err := tronClient.Client.Transfer(tronClient.Address, toAddress, transferAmount.Int64())
	require.NoError(t, err)
	fmt.Printf("refBlockBytes: %x\n", tx.Transaction.RawData.RefBlockBytes)
	fmt.Printf("refBlockHash: %x\n", tx.Transaction.RawData.RefBlockHash)
}

func TestBlockHash(t *testing.T) {
	nodeURL := TronTestNet
	tronClient, err := NewTronClient(nodeURL)
	if err != nil {
		log.Fatalf("failed to initialize Tron client: %v", err)
	}
	defer tronClient.Client.Stop()

	block, err := tronClient.Client.GetNowBlock()
	require.NoError(t, err)

	fmt.Printf("Block Number: %d\n", block.GetBlockHeader().GetRawData().Number)
	fmt.Printf("Block Hash: %x\n", block.GetBlockid())

	calculatedHash, err := GetBlockHash(block.GetBlockHeader())
	require.NoError(t, err)
	fmt.Printf("Calculated Block Hash: %s\n", calculatedHash)

	fmt.Printf("Parent Block Hash: %x\n", block.GetBlockHeader().GetRawData().ParentHash)
}

func TestABI(t *testing.T) {
	swapId := big.NewInt(1234567890)
	fromToken := "TPpxTLNp5XCGTgqnY5jxijhRF4aYqfQN12"
	signature := []byte("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	method := "swap(uint256,address,bytes)"
	jsonString := fmt.Sprintf(`
	[
		{"uint256": "%s"},
		{"address": "%s"},
		{"bytes": "%s"}
	]
	`, swapId, fromToken, hex.EncodeToString(signature))

	param, err := abi.LoadFromJSON(jsonString)
	require.NoError(t, err)
	dataBytes, err := abi.Pack(method, param)
	require.NoError(t, err)

	fmt.Printf("Packed Data: %x\n", dataBytes)
}
