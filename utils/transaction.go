package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/protobuf/proto"
)

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
