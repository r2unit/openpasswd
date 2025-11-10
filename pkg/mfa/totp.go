package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type TOTPKey struct {
	Secret      string
	Issuer      string
	AccountName string
}

func GenerateTOTPSecret(accountName string) (*TOTPKey, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	encodedSecret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)

	return &TOTPKey{
		Secret:      encodedSecret,
		Issuer:      "OpenPasswd",
		AccountName: accountName,
	}, nil
}

func (k *TOTPKey) URL() string {
	v := url.Values{}
	v.Set("secret", k.Secret)
	v.Set("issuer", k.Issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	v.Set("period", "30")

	label := url.PathEscape(k.Issuer + ":" + k.AccountName)
	return fmt.Sprintf("otpauth://totp/%s?%s", label, v.Encode())
}

func ValidateTOTP(secret string, code string) bool {
	expectedCode := generateTOTP(secret, time.Now().Unix()/30)
	return code == expectedCode
}

func generateTOTP(secret string, counter int64) string {
	secret = strings.ToUpper(secret)
	secret = strings.ReplaceAll(secret, " ", "")

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	code := truncated % 1000000

	return fmt.Sprintf("%06d", code)
}

func GenerateQRCodeASCII(key *TOTPKey) (string, error) {
	return "", nil
}


func GetTOTPCode() (string, error) {
	fmt.Print("Enter 6-digit TOTP code: ")
	var code string
	fmt.Scanln(&code)
	return code, nil
}

func EncodeTOTPSecret(secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(secret))
}

func DecodeTOTPSecret(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
