package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"lproxy/servercfg"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// one month
	myTimeExpired = (30 * 24 * 60 * 60)
)

const (
	errTokenSuccess = 0
	errTokenEmpty   = 1
	errTokenDecrypt = 2
	errTokenFormat  = 3
	errTokenExpired = 4
)

func verifyToken(r *http.Request) (string, bool) {
	var tk = r.Header.Get("tk")

	if tk == "" {
		return "", false
	}

	v, e := parseTK(tk)
	if e == errTokenSuccess {
		return v, true
	}

	return v, false
}

// GenTK 生成一个加密的token
func GenTK(account string) string {
	var plainTK = fmt.Sprintf("%s@%d", account, time.Now().Unix())
	// log.Println("GenTK, plainTK is:", plainTK)
	return encrypt([]byte(servercfg.TokenKey), plainTK)
}

func parseTK(token string) (string, int) {
	// log.Printf("ParseTk, tok:%s, len:%d\n", token, len(token))
	if token == "" {
		return "", errTokenEmpty
	}

	var plainTK, err = decrypt([]byte(servercfg.TokenKey), token)
	if err != nil {
		log.Println("ParseTK, err:", err)
		return "", errTokenDecrypt
	}

	//log.Println("ParseTK, plainTK is:", plainTK)

	var splits = strings.Split(plainTK, "@")
	if len(splits) != 2 {
		log.Println("ParseTK, err: no @ at text")
		return "", errTokenFormat
	}

	timestamp, err := strconv.Atoi(splits[1])
	if err != nil {
		log.Println("ParseTK, err: ", err)
		return "", errTokenFormat
	}

	var now = int(time.Now().Unix())
	//log.Printf("ParseTK, account:%s, timestamp:%d, now:%d", splits[0], timestamp, now)

	if now-timestamp > (myTimeExpired) {
		log.Println("ParseTK, token has been expired")
		return "", errTokenExpired
	}

	return splits[0], errTokenSuccess
}

// encrypt string to base64 crypto using AES
func encrypt(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func decrypt(key []byte, cryptoText string) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), nil
}
