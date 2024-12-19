package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	SECRET_KEY string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	} else {
		log.Println(".env file successfully loaded")
	}

	SECRET_KEY = os.Getenv("SECRET_KEY_WEBHOOK")

	if SECRET_KEY == "" {
		log.Println("SECRET_KEY is not set")
		os.Exit(1)
	}
}

const (
	PORT = 4000
)

// Mapear o algoritmo a partir do cabeçalho
var algorithmMap = map[string]func() hash.Hash{
	"HMAC_SHA1_HEX":   sha1.New,
	"HMAC_SHA256_HEX": sha256.New,
	"HMAC_SHA512_HEX": sha512.New,
}

// Função para verificar a assinatura do webhook
func verifySignature(req *http.Request, body []byte) bool {
	digest := req.Header.Get("X-Payload-Digest")
	digestAlg := req.Header.Get("X-Payload-Digest-Alg")

	log.Println("Digest:", digest)
	log.Println("Digest Algorithm:", digestAlg)

	if digest == "" || digestAlg == "" {
		log.Println("Cabeçalhos de digest ou algoritmo ausentes")
		return false
	}

	// Obter a função de hash
	hashFunc, ok := algorithmMap[digestAlg]
	if !ok {
		log.Println("Algoritmo de digest não suportado")
		return false
	}

	// Calcular o digest
	h := hmac.New(hashFunc, []byte(SECRET_KEY))
	h.Write(body)
	calculatedDigest := hex.EncodeToString(h.Sum(nil))

	log.Println("Calculated Digest:", calculatedDigest)
	log.Println("Received Digest:", digest)

	return calculatedDigest == digest
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Erro ao ler o corpo da requisição:", err)
		http.Error(w, "Erro ao processar webhook", http.StatusBadRequest)
		return
	}

	if !verifySignature(r, body) {
		log.Println("Assinatura do webhook inválida")
		http.Error(w, "Assinatura inválida", http.StatusUnauthorized)
		return
	}

	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Println("Erro ao decodificar JSON do webhook:", err)
		http.Error(w, "Erro ao processar webhook", http.StatusBadRequest)
		return
	}

	log.Println("Webhook recebido:", event)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook recebido com sucesso"))
}

func main() {
	http.HandleFunc("/", webhookHandler)
	log.Printf("Servidor de webhook rodando em http://localhost:%d/", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil))
}
