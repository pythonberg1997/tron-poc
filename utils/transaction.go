package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type RefBlockInfo struct {
	RefBlockBytes []byte
	RefBlockHash  []byte
}

func CreateTransferTransaction(fromAddress, toAddress string, amount int64, refBlock *RefBlockInfo) (*core.Transaction, error) {
	fromAddr, err := address.Base58ToAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %v", err)
	}

	toAddr, err := address.Base58ToAddress(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %v", err)
	}

	transferContract := &core.TransferContract{
		OwnerAddress: fromAddr.Bytes(),
		ToAddress:    toAddr.Bytes(),
		Amount:       amount,
	}

	contractAny, err := anypb.New(transferContract)
	if err != nil {
		return nil, fmt.Errorf("failed to create any for transfer contract: %v", err)
	}

	contract := &core.Transaction_Contract{
		Type:      core.Transaction_Contract_TransferContract,
		Parameter: contractAny,
	}

	now := time.Now()
	expiration := now.Add(1 * time.Minute).UnixMilli()

	rawData := &core.TransactionRaw{
		RefBlockBytes: refBlock.RefBlockBytes,
		RefBlockHash:  refBlock.RefBlockHash,
		Expiration:    expiration,
		Timestamp:     now.UnixMilli(),
		Contract:      []*core.Transaction_Contract{contract},
	}

	transaction := &core.Transaction{
		RawData: rawData,
	}

	return transaction, nil
}

func CreateTriggerContractTransaction(
	fromAddress, contractAddress, method, jsonString string,
	callValue, feeLimit int64,
	refBlock *RefBlockInfo,
) (*core.Transaction, error) {
	fromAddr, err := address.Base58ToAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %v", err)
	}

	contractAddr, err := address.Base58ToAddress(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %v", err)
	}

	triggerContract := &core.TriggerSmartContract{
		OwnerAddress:    fromAddr.Bytes(),
		ContractAddress: contractAddr.Bytes(),
		CallValue:       callValue,
	}

	if method != "" {
		param, err := abi.LoadFromJSON(jsonString)
		if err != nil {
			return nil, err
		}

		dataBytes, err := abi.Pack(method, param)
		if err != nil {
			return nil, err
		}

		triggerContract.Data = dataBytes
	}

	contractAny, err := anypb.New(triggerContract)
	if err != nil {
		return nil, fmt.Errorf("failed to create any for trigger contract: %v", err)
	}

	contract := &core.Transaction_Contract{
		Type:      core.Transaction_Contract_TriggerSmartContract,
		Parameter: contractAny,
	}

	now := time.Now()
	expiration := now.Add(1 * time.Minute).UnixMilli()

	rawData := &core.TransactionRaw{
		RefBlockBytes: refBlock.RefBlockBytes,
		RefBlockHash:  refBlock.RefBlockHash,
		Expiration:    expiration,
		Timestamp:     now.UnixMilli(),
		Contract:      []*core.Transaction_Contract{contract},
		FeeLimit:      feeLimit,
	}

	transaction := &core.Transaction{
		RawData: rawData,
	}

	return transaction, nil
}

func SignTransactions(transaction *core.Transaction, privateKey *ecdsa.PrivateKey) (*core.Transaction, error) {
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		return nil, err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}
	transaction.Signature = append(transaction.Signature, signature)
	return transaction, nil
}

func GetTransactionHash(transaction *core.Transaction) (string, error) {
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		return "", err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	return hex.EncodeToString(hash), nil
}

func GetBlockHash(block *core.BlockHeader) (string, error) {
	rawData, err := proto.Marshal(block.GetRawData())
	if err != nil {
		return "", err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	return hex.EncodeToString(hash), nil
}
