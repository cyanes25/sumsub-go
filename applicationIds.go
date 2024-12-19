package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	BASE_URL = "https://api.sumsub.com"
)

var (
	APP_TOKEN  string
	SECRET_KEY string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	} else {
		log.Println(".env file successfully loaded")
	}

	APP_TOKEN = os.Getenv("APP_TOKEN")
	SECRET_KEY = os.Getenv("SECRET_KEY")

	log.Printf("APP_TOKEN: %s", APP_TOKEN)
	log.Printf("SECRET_KEY: %s", SECRET_KEY)
}

func createSignature(method, path string, timestamp int64) string {
	message := fmt.Sprintf("%d%s%s", timestamp, method, path)
	h := hmac.New(sha256.New, []byte(SECRET_KEY))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func listApplicants() {
	if APP_TOKEN == "" || SECRET_KEY == "" {
		log.Fatalf("APP_TOKEN or SECRET_KEY is not set")
	}

	method := "GET"
	path := "/resources/applicants/-/levels"
	timestamp := time.Now().Unix()
	signature := createSignature(method, path, timestamp)

	headers := map[string]string{
		"X-App-Token":      APP_TOKEN,
		"X-App-Access-Sig": signature,
		"X-App-Access-Ts":  fmt.Sprintf("%d", timestamp),
		"Accept":           "application/json",
	}

	url := BASE_URL + path

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("Erro ao criar a requisição: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler a resposta: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			log.Fatalf("Erro ao decodificar JSON: %v", err)
		}
		formattedJSON, _ := json.MarshalIndent(data, "", "  ")
		log.Println("List of Applicants:", string(formattedJSON))
	} else {
		log.Printf("Erro na requisição: %s\n%s", resp.Status, string(body))
	}
}

func main() {
	listApplicants()
}
