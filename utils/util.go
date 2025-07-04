package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

func LoadPrivateKeyFromEnv() (*ecdsa.PrivateKey, error) {
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
