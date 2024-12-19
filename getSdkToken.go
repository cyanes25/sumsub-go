package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	SUMSUB_APP_TOKEN  string
	SUMSUB_SECRET_KEY string
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	} else {
		fmt.Println(".env file successfully loaded")
	}

	SUMSUB_APP_TOKEN = os.Getenv("APP_TOKEN")
	SUMSUB_SECRET_KEY = os.Getenv("SECRET_KEY")

	if SUMSUB_APP_TOKEN == "" || SUMSUB_SECRET_KEY == "" {
		fmt.Println("Error: SUMSUB_APP_TOKEN or SUMSUB_SECRET_KEY is not set")
		os.Exit(1)
	}
}

const (
	SUMSUB_BASE_URL    = "https://api.sumsub.com"
	DEFAULT_LEVEL_NAME = "basic-kyb-BRLA"
)

func createSignature(method, endpoint string, timestamp int64, body string) string {
	message := fmt.Sprintf("%d%s%s", timestamp, strings.ToUpper(method), endpoint)
	if body != "" {
		message += body
	}
	h := hmac.New(sha256.New, []byte(SUMSUB_SECRET_KEY))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func createAccessToken(externalUserId, levelName string, ttlInSecs int) (*http.Request, error) {
	timestamp := time.Now().Unix()
	endpoint := fmt.Sprintf("/resources/accessTokens?userId=%s&ttlInSecs=%d&levelName=%s", url.QueryEscape(externalUserId), ttlInSecs, url.QueryEscape(levelName))
	fullURL := SUMSUB_BASE_URL + endpoint

	signature := createSignature("POST", endpoint, timestamp, "")

	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-App-Token", SUMSUB_APP_TOKEN)
	req.Header.Set("X-App-Access-Ts", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-App-Access-Sig", signature)

	return req, nil
}

func main() {
	externalUserId := fmt.Sprintf("random-GoToken-%s", randomString(9))
	levelName := DEFAULT_LEVEL_NAME
	ttlInSecs := 1200

	fmt.Println("External UserID:", externalUserId)

	req, err := createAccessToken(externalUserId, levelName, ttlInSecs)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Response:", string(body))
	} else {
		fmt.Println("Error:", string(body))
	}
}

func randomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
