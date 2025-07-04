package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"tron-poc/utils"
)

const (
	TronMainNet = "grpc.trongrid.io:50051"
	TronTestNet = "grpc.shasta.trongrid.io:50051"
)

type TronClient struct {
	client     *client.GrpcClient
	privateKey *ecdsa.PrivateKey
	address    string
}

func loadPrivateKeyFromEnv() (*ecdsa.PrivateKey, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	privateKeyHex := os.Getenv("PRIVATE_KEY")
	privateKeyBytes, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %v", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %v", err)
	}

	return privateKey, nil
}

func NewTronClient(nodeURL string) (*TronClient, error) {
	os.Setenv("GOTRON_SDK_DEBUG", "true")
	os.Setenv("TRON_NODE_TLS", "true")

	conn := client.NewGrpcClient(nodeURL)
	err := conn.Start(grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %v", err)
	}

	privateKey, err := loadPrivateKeyFromEnv()
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	tronAddress := address.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).String()
	log.Println("eth address", address.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).Hex())

	return &TronClient{
		client:     conn,
		privateKey: privateKey,
		address:    tronAddress,
	}, nil
}

func (tc *TronClient) GetBalance() (*big.Int, error) {
	account, err := tc.client.GetAccount(tc.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account information: %v", err)
	}

	return big.NewInt(account.Balance), nil
}

func (tc *TronClient) CreateTransfer(toAddress string, amount *big.Int) (string, error) {
	tx, err := tc.client.Transfer(tc.address, toAddress, amount.Int64())
	if err != nil {
		return "", fmt.Errorf("failed to create transfer transaction: %v", err)
	}

	return tc.SendTx(tx)
}

func (tc *TronClient) DeployContract(bytecode string) (string, error) {
	tx, err := tc.client.DeployContract(tc.address, "POC", nil, bytecode, 10000000, 100, 100)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %v", err)
	}

	return tc.SendTx(tx)
}

func (tc *TronClient) SendTx(tx *api.TransactionExtention) (string, error) {
	signedTx, err := utils.SignTransactions(tx.Transaction, tc.privateKey)
	if err != nil {
		return "", err
	}

	result, err := tc.client.Broadcast(signedTx)
	if err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("broadcast failed: %s", result.Message)
	}

	return hex.EncodeToString(tx.Txid), nil
}

func main() {
	fmt.Println("=== Tron POC tool ===")

	nodeURL := TronTestNet
	tronClient, err := NewTronClient(nodeURL)
	if err != nil {
		log.Fatalf("failed to initialize Tron client: %v", err)
	}
	defer tronClient.client.Stop()

	fmt.Printf("sender address: %s\n", tronClient.address)

	fmt.Println("\n=== check balance ===")
	balance, err := tronClient.GetBalance()
	if err != nil {
		fmt.Printf("failed to get balance: %v\n", err)
		return
	}

	balanceTRX := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1000000))
	fmt.Printf("current balance: %s TRX\n", balanceTRX.String())

	// fmt.Println("\n=== deploy contract ===")
	// bytecode := "0x6080806040523460145760ba90816100198239f35b5f80fdfe60808060405260043610156011575f80fd5b5f3560e01c9081632f048afa14606e5781633fa4f24514605757506357de26a4146039575f80fd5b346053575f36600319011260535760205f54604051908152f35b5f80fd5b346053575f3660031901126053576020905f548152f35b3460535760203660031901126053576004355f5500fea26474726f6e582212204d8543cfa16ba3281148633fc0d4c9693a1b6cbd01f92d05d03d133391b3e7cb64736f6c63430008170033"

	// // TAQo6ZE1ozpmc2oj4uZ2XRHdtUzWDqRQu6
	// txHash, err := tronClient.DeployContract(strings.TrimSpace(bytecode))
	// if err != nil {
	// 	fmt.Printf("failed to deploy contract: %v\n", err)
	// 	return
	// }
	// fmt.Printf("contract deployed successfully, txHash: %s\n", txHash)

	fmt.Println("\n=== read contract ===")
	result, err := tronClient.client.TriggerConstantContract(
		"",
		"TAQo6ZE1ozpmc2oj4uZ2XRHdtUzWDqRQu6",
		"read()",
		"",
	)
	if err != nil {
		fmt.Printf("failed to call contract: %v\n", err)
		return
	}
	fmt.Printf("result: %d\n", result.ConstantResult[0])

	fmt.Println("\n=== write contract ===")
	tx, err := tronClient.client.TriggerContract(
		tronClient.address,
		"TAQo6ZE1ozpmc2oj4uZ2XRHdtUzWDqRQu6",
		"write(uint256)",
		`[{"uint256": "1"}]`,
		10000000,
		0,
		"",
		0,
	)
	if err != nil {
		fmt.Printf("failed to create contract call tx: %v\n", err)
		return
	}
	txHash, err := tronClient.SendTx(tx)
	if err != nil {
		fmt.Printf("failed to send contract call tx: %v\n", err)
		return
	}
	fmt.Printf("contract call tx sent successfully, txHash: %s\n", txHash)

	// fmt.Println("\n=== transfer ===")
	// toAddress := "TGj1Ej1qRzL9feLTLhjwgxXF4Ct6GTWg2U"
	// transferAmount := big.NewInt(100000)
	//
	// fmt.Printf("recipient address: %s\n", toAddress)
	// fmt.Printf("transfer amount: %s TRX\n", new(big.Float).Quo(new(big.Float).SetInt(transferAmount), big.NewFloat(1000000)).String())
	//
	// txHash, err := tronClient.CreateTransfer(toAddress, transferAmount)
	// if err != nil {
	// 	fmt.Printf("failed to create transfer: %v\n", err)
	// 	return
	// }
	//
	// fmt.Printf("txHash: %s\n", txHash)
	//
	// fmt.Println("\n=== check balance after transfer ===")
	// newBalance, err := tronClient.GetBalance()
	// if err != nil {
	// 	fmt.Printf("failed to get balance after transfer: %v\n", err)
	// 	return
	// }
	//
	// newBalanceTRX := new(big.Float).Quo(new(big.Float).SetInt(newBalance), big.NewFloat(1000000))
	// fmt.Printf("new balance: %s TRX\n", newBalanceTRX.String())
}
