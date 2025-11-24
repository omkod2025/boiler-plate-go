package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func main() {
	// สร้าง RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("Failed to generate RSA key pair:", err)
	}

	// สร้าง private key PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// สร้าง public key PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal("Failed to marshal public key:", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Encode เป็น base64
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyPEM)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyPEM)

	// เขียน private key ลงไฟล์
	err = os.WriteFile("private_key.pem", privateKeyPEM, 0600)
	if err != nil {
		log.Fatal("Failed to write private key:", err)
	}

	// เขียน public key ลงไฟล์
	err = os.WriteFile("public_key.pem", publicKeyPEM, 0644)
	if err != nil {
		log.Fatal("Failed to write public key:", err)
	}

	// เขียน base64 encoded keys ลงไฟล์
	err = os.WriteFile("private_key_base64.txt", []byte(privateKeyBase64), 0600)
	if err != nil {
		log.Fatal("Failed to write base64 private key:", err)
	}

	err = os.WriteFile("public_key_base64.txt", []byte(publicKeyBase64), 0644)
	if err != nil {
		log.Fatal("Failed to write base64 public key:", err)
	}

	fmt.Println("RSA key pair generated and encoded successfully!")
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Println("- private_key.pem (PEM format)")
	fmt.Println("- public_key.pem (PEM format)")
	fmt.Println("- private_key_base64.txt (Base64 encoded)")
	fmt.Println("- public_key_base64.txt (Base64 encoded)")
	fmt.Println()
	fmt.Println("For .env file, use the base64 encoded values:")
	fmt.Println("JWT_PRIVATE_KEY=" + privateKeyBase64)
	fmt.Println("JWT_PUBLIC_KEY=" + publicKeyBase64)
	fmt.Println()
	fmt.Println("Or copy from the base64 files:")
	fmt.Println("JWT_PRIVATE_KEY=$(cat private_key_base64.txt)")
	fmt.Println("JWT_PUBLIC_KEY=$(cat public_key_base64.txt)")
	fmt.Println()
	fmt.Println("Add these to your .env file:")
	fmt.Println("JWT_PRIVATE_KEY=$(cat private_key_base64.txt)")
	fmt.Println("JWT_PUBLIC_KEY=$(cat public_key_base64.txt)")
}
