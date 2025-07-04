package main

import (
	"fmt"
	"log"
	"math/big"

	"tron-poc/utils"
)

func main() {
	fmt.Println("=== Tron POC tool ===")

	nodeURL := utils.TronTestNet
	tronClient, err := utils.NewTronClient(nodeURL)
	if err != nil {
		log.Fatalf("failed to initialize Tron client: %v", err)
	}
	defer tronClient.Client.Stop()

	fmt.Printf("sender address: %s\n", tronClient.Address)

	fmt.Println("\n=== check balance ===")
	balance, err := tronClient.GetBalance()
	if err != nil {
		fmt.Printf("failed to get balance: %v\n", err)
		return
	}

	balanceTRX := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1000000))
	fmt.Printf("current balance: %s TRX\n", balanceTRX.String())

	refBlock, err := tronClient.GetRefBlockInfo()
	if err != nil {
		fmt.Printf("failed to get ref block info: %v\n", err)
		return
	}

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
	result, err := tronClient.Client.TriggerConstantContract(
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
	parameter := `
	[
		{"uint256": "2"}
	]
	`

	txHash, err := tronClient.CreateTriggerContractOffline(
		"TAQo6ZE1ozpmc2oj4uZ2XRHdtUzWDqRQu6",
		"write(uint256)",
		parameter,
		0,        // callValue
		10000000, // feeLimit
		refBlock,
	)
	if err != nil {
		fmt.Printf("failed to create contract call tx: %v\n", err)
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
	// txHash, err := tronClient.CreateTransferOffline(toAddress, transferAmount, refBlock)
	// if err != nil {
	// 	fmt.Printf("failed to create transfer with external refBlock: %v\n", err)
	// 	return
	// }
	// fmt.Printf("transfer with external refBlock created, txHash: %s\n", txHash)
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
