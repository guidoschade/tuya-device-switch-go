package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

var ( Token string )
var device_id = ""
var code_var = ""
var value_var = ""

var host = ""
var clientId = ""
var secret = ""
var mode = ""

type TokenResponse struct {
	Result struct {
		AccessToken  string `json:"access_token"`
		ExpireTime   int    `json:"expire_time"`
		RefreshToken string `json:"refresh_token"`
		UID          string `json:"uid"`
	} `json:"result"`
	Success bool  `json:"success"`
	T       int64 `json:"t"`
}

// main function
func main() {

	if len(os.Args[1:]) < 1 {
		print("Error, required command line arguments - use '-h' for more information\n")
		os.Exit(1)
	}

	// reading command line flags (if any)
	flag.StringVar(&device_id, "d", "x", "DeviceID - acquired from Tuya app or developer portal")
	flag.StringVar(&code_var, "c", "switch_1", "Device Code")
	flag.StringVar(&value_var, "v", "true", "Device Value to set")
	flag.StringVar(&host, "H", "https://openapi.tuyaeu.com", "Host")
	flag.StringVar(&clientId, "i", "x", "ClientID - Tuya Client ID")
	flag.StringVar(&secret, "s", "x", "Secret - Tuya Client Secret")
	flag.StringVar(&mode, "m", "set", "mode - view (info) / set")
	flag.Parse()

	// getting token
	GetToken()

	// show device status or switch / send command
        if mode == "view" {
	  GetDevice(device_id)
        } else {
	  SendCommand(device_id)
	}
}

// getting token from Tuya
func GetToken() {
	method := "GET"
	body := []byte(``)
	req, _ := http.NewRequest(method, host+"/v1.0/token?grant_type=1", bytes.NewReader(body))

	buildHeader(req, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	ret := TokenResponse{}
	json.Unmarshal(bs, &ret)

	if v := ret.Result.AccessToken; v != "" {
		Token = v
	}
}

// sending single command
func SendCommand(deviceId string) {
	method := "POST"
	body := []byte(fmt.Sprintf(`{"commands": [{"code": "%s","value": %s}]}`, code_var, value_var))
	req, _ := http.NewRequest(
		method,
		host+"/v1.0/iot-03/devices/"+deviceId+"/commands",
		bytes.NewReader(body),
	)

	buildHeader(req, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(bs))
}

// getting device from ID
func GetDevice(deviceId string) {
	method := "GET"
	body := []byte(``)
	req, _ := http.NewRequest(method, host+"/v1.0/devices/"+deviceId, bytes.NewReader(body))

	buildHeader(req, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(bs))
}

func buildHeader(req *http.Request, body []byte) {
	req.Header.Set("client_id", clientId)
	req.Header.Set("sign_method", "HMAC-SHA256")

	ts := fmt.Sprint(time.Now().UnixNano() / 1e6)
	req.Header.Set("t", ts)

	if Token != "" {
		req.Header.Set("access_token", Token)
	}

	sign := buildSign(req, body, ts)
	req.Header.Set("sign", sign)
}

func buildSign(req *http.Request, body []byte, t string) string {
	headers := getHeaderStr(req)
	urlStr := getUrlStr(req)
	contentSha256 := Sha256(body)
	stringToSign := req.Method + "\n" + contentSha256 + "\n" + headers + "\n" + urlStr
	signStr := clientId + Token + t + stringToSign
	sign := strings.ToUpper(HmacSha256(signStr, secret))
	return sign
}

func Sha256(data []byte) string {
	sha256Contain := sha256.New()
	sha256Contain.Write(data)
	return hex.EncodeToString(sha256Contain.Sum(nil))
}

func getUrlStr(req *http.Request) string {
	url := req.URL.Path
	keys := make([]string, 0, 10)

	query := req.URL.Query()
	for key, _ := range query {
		keys = append(keys, key)
	}
	if len(keys) > 0 {
		url += "?"
		sort.Strings(keys)
		for _, keyName := range keys {
			value := query.Get(keyName)
			url += keyName + "=" + value + "&"
		}
	}

	if url[len(url)-1] == '&' {
		url = url[:len(url)-1]
	}
	return url
}

func getHeaderStr(req *http.Request) string {
	signHeaderKeys := req.Header.Get("Signature-Headers")
	if signHeaderKeys == "" {
		return ""
	}
	keys := strings.Split(signHeaderKeys, ":")
	headers := ""
	for _, key := range keys {
		headers += key + ":" + req.Header.Get(key) + "\n"
	}
	return headers
}

func HmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
