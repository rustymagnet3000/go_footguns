// Command main demonstrates a Go-specific security footgun: a security check
// whose only failure signal is an error silently fails OPEN when that error is
// discarded. Run it and watch a forged message get "processed".
//
//	go run ./04-error-fail-open
package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	msg := []byte("transfer £10 to alice")
	sum := sha256.Sum256(msg)
	sig, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, sum[:])
	if err != nil {
		panic(err)
	}

	// An attacker swaps the signed message for a different one.
	tampered := []byte("transfer £9000 to mallory")
	tsum := sha256.Sum256(tampered)

	// FOOTGUN: VerifyPKCS1v15 reports an invalid signature ONLY via its error
	// return. Called as a bare statement, that error is discarded — so the
	// forged message sails straight through. This is fail-OPEN.
	rsa.VerifyPKCS1v15(&priv.PublicKey, crypto.SHA256, tsum[:], sig)
	fmt.Println("FAIL-OPEN   → processed:", string(tampered))

	// CORRECT: the error IS the check. Handle it, and the forged message stops here.
	if err := rsa.VerifyPKCS1v15(&priv.PublicKey, crypto.SHA256, tsum[:], sig); err != nil {
		fmt.Println("FAIL-CLOSED → rejected:", err)
		return
	}
	fmt.Println("processed:", string(tampered))
}
