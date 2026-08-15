package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// Main script entry point to generate 2048-bit RSA key pair for local JWT signing and verification.
func main() {
	keyDir := "./keys"
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		fmt.Printf("Error creating keys directory: %v\n", err)
		os.Exit(1)
	}

	privPath := filepath.Join(keyDir, "private.pem")
	pubPath := filepath.Join(keyDir, "public.pem")

	if _, err := os.Stat(privPath); err == nil {
		fmt.Println("RSA key pair already exists in ./keys. Skipping generation.")
		return
	}

	fmt.Println("Generating 2048-bit RSA key pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate RSA private key: %v\n", err)
		os.Exit(1)
	}

	// Save Private Key (PKCS1)
	privFile, err := os.Create(privPath)
	if err != nil {
		fmt.Printf("Failed to create private key file: %v\n", err)
		os.Exit(1)
	}
	defer privFile.Close()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		fmt.Printf("Failed to write private key PEM: %v\n", err)
		os.Exit(1)
	}

	// Save Public Key (PKIX)
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Printf("Failed to marshal public key: %v\n", err)
		os.Exit(1)
	}

	pubFile, err := os.Create(pubPath)
	if err != nil {
		fmt.Printf("Failed to create public key file: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()

	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		fmt.Printf("Failed to write public key PEM: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated RSA 2048-bit key pair in ./keys/ (private.pem, public.pem)")
}
