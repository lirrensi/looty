// FILE: internal/certgen/certgen.go
// PURPOSE: Generate an in-memory self-signed TLS certificate per server run with a human-readable friend code.
// OWNS: TLS certificate generation, friend code creation, fingerprint computation.
// EXPORTS: GenerateSelfSigned
// DOCS: agent_chat/plan_tls-paradigm_2026-05-17.md

package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"strings"
	"math/big"
	"time"
)

var adjectives = []string{
	"brave", "calm", "eager", "fancy", "gentle", "happy", "jolly", "kind",
	"lively", "merry", "nice", "proud", "silly", "witty", "zealous", "cosmic",
	"sunny", "rusty", "swift", "mighty",
}

var nouns = []string{
	"apple", "beach", "cactus", "dolphin", "eagle", "falcon", "garden",
	"harbor", "island", "jungle", "kite", "lemon", "meadow", "nebula",
	"ocean", "panda", "quartz", "river", "star", "tiger",
}

func randomWord(list []string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(list))))
	if err != nil {
		// Fallback to a deterministic choice if crypto/rand fails
		return list[0]
	}
	return list[n.Int64()]
}

func randomFourDigits() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "0000"
	}
	return fmt.Sprintf("%04d", n.Int64())
}

func generateFriendCode() string {
	adj := randomWord(adjectives)
	noun := randomWord(nouns)
	digits := randomFourDigits()
	return fmt.Sprintf("looty-%s-%s-%s", adj, noun, digits)
}

// GenerateSelfSigned creates a new 2048-bit RSA self-signed certificate valid for 24 hours.
// It returns the tls.Certificate, the SHA-256 fingerprint (colon-separated hex), the friend code, and any error.
func GenerateSelfSigned() (*tls.Certificate, string, string, error) {
	friendCode := generateFriendCode()

	// 1. Generate 2048-bit RSA key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// 2. Build certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: friendCode,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{"localhost", "looty.local"},
		BasicConstraintsValid: true,
	}

	// 3. Sign the certificate with the same private key
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// 4. Build tls.Certificate
	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}

	// 5. Compute SHA-256 fingerprint of DER-encoded cert
	hash := sha256.Sum256(certDER)
	fingerprint := formatFingerprint(hash[:])

	return &cert, fingerprint, friendCode, nil
}

func formatFingerprint(hash []byte) string {
	hexStr := strings.ToUpper(hex.EncodeToString(hash))
	var result []byte
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			result = append(result, ':')
		}
		result = append(result, hexStr[i], hexStr[i+1])
	}
	return string(result)
}
