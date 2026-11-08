// signrelease checksums.txt'yi güncelleme anahtarıyla imzalar (scripts/build-release.sh).
//
//	UPDATE_SIGNING_KEY=<base64 tohum> signrelease dist/checksums.txt   → dist/checksums.txt.sig
//	signrelease -gen <anahtar dosyası>                                 → yeni anahtar, açık yarısını yazar
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "signrelease:", msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) == 3 && os.Args[1] == "-gen" {
		if _, err := os.Stat(os.Args[2]); err == nil {
			fail("anahtar dosyası zaten var, üzerine yazılmaz")
		}
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fail(err.Error())
		}
		if err := os.WriteFile(os.Args[2], []byte(base64.StdEncoding.EncodeToString(priv.Seed())), 0600); err != nil {
			fail(err.Error())
		}
		fmt.Println(base64.StdEncoding.EncodeToString(pub))
		return
	}
	if len(os.Args) != 2 {
		fail("kullanım: signrelease <dosya>")
	}
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(os.Getenv("UPDATE_SIGNING_KEY")))
	if err != nil || len(seed) != ed25519.SeedSize {
		fail("UPDATE_SIGNING_KEY geçersiz")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fail(err.Error())
	}
	sig := ed25519.Sign(ed25519.NewKeyFromSeed(seed), data)
	if err := os.WriteFile(os.Args[1]+".sig", []byte(base64.StdEncoding.EncodeToString(sig)+"\n"), 0644); err != nil {
		fail(err.Error())
	}
}
