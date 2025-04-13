package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"

	"github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/security/signer"
)

const keyDir = "/opt/worker_keychain"

func LoadOrCreateECDSASigner() ndn.Signer {
	privPath := filepath.Join(keyDir, "ecdsa.key")

	var privKey *ecdsa.PrivateKey

	// Load key if it exists
	if _, err := os.Stat(privPath); err == nil {
		data, err := os.ReadFile(privPath)
		if err != nil {
			log.Fatalf("Failed to read ECDSA key: %v", err)
		}
		block, _ := pem.Decode(data)
		if block == nil {
			log.Fatalf("Invalid PEM in key file")
		}
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			log.Fatalf("Failed to parse ECDSA key: %v", err)
		}
		privKey = key
	} else {
		// Generate new key
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalf("Failed to generate ECDSA key: %v", err)
		}
		privKey = key

		// Save as PEM
		keyBytes, _ := x509.MarshalECPrivateKey(privKey)
		block := pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
		os.MkdirAll(keyDir, 0700)
		err = os.WriteFile(privPath, pem.EncodeToMemory(&block), 0600)
		if err != nil {
			log.Fatalf("Failed to save ECDSA key: %v", err)
		}
		log.Printf("Generated new ECDSA key at %s", privPath)
	}

	sigName, err := encoding.NameFromStr("/local/worker/ecdsa-key")
	if err != nil {
		log.Fatalf("Failed to generate signature name: %v", err)
	}

	return signer.NewEccSigner(sigName, privKey)
}
