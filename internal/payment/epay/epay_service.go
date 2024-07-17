package epay

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Service struct {
	client *http.Client
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string `json:"scope"`
}

type PaymentResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"`
	Message   string  `json:"message"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	InvoiceID string  `json:"invoice_id"`
}

var (
	ClientID           = "test"
	ClientSecret       = "yF587AV9Ms94qN2QShFzVR3vFnWkhjbAK3sG"
	DefaultPaymentData = `{
 		"hpan":"4405639704015096",
		"expDate":"0125",
		"cvc":"815",
		"terminalId":
		"67e34d63-102f-4bd1-898e-370781d0074d"
	}`
)

func NewService() *Service {
	return &Service{
		client: http.DefaultClient,
	}
}

func (s *Service) Token() (string, error) {
	tokenURL := "https://testoauth.homebank.kz/epay2/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "webapi usermanagement email_send verification statement statistics payment")
	data.Set("client_id", ClientID)
	data.Set("client_secret", ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get token, status code: " + resp.Status)
	}

	token := TokenResponse{}

	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func (s *Service) encryptData(data string) (string, error) {
	url := "https://testepay.homebank.kz/api/public.rsa"

	resp, err := s.client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(body)
	if block == nil {
		return "", errors.New("failed to decode pem block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub.(*rsa.PublicKey), []byte(data), nil)
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), []byte(data))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (s *Service) Pay(token string) (PaymentResponse, error) {
	payURL := "https://testepay.homebank.kz/api/payment/cryptopay"

	encryptedData, err := s.encryptData(DefaultPaymentData)
	if err != nil {
		return PaymentResponse{}, err
	}

	requestData := map[string]interface{}{
		"amount":          100,
		"currency":        "KZT",
		"name":            "JON JONSON",
		"cryptogram":      encryptedData,
		"invoiceId":       "000001",
		"invoiceIdAlt":    "8564546",
		"description":     "test payment",
		"accountId":       "uuid000001",
		"email":           "jj@example.com",
		"phone":           "77777777777",
		"cardSave":        true,
		"data":            `{"statement":{"name":"Arman Ali","invoiceID":"80000016"}}`,
		"postLink":        "https://testmerchant/order/1123",
		"failurePostLink": "https://testmerchant/order/1123/fail",
	}

	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return PaymentResponse{}, err
	}

	req, err := http.NewRequest("POST", payURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return PaymentResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return PaymentResponse{}, err
	}
	defer resp.Body.Close()

	data := json.RawMessage{}
	json.NewDecoder(resp.Body).Decode(&data)
	fmt.Println(string(data)) // error code 327 for some reason

	if resp.StatusCode != http.StatusOK {
		return PaymentResponse{}, errors.New("failed to pay, status code: " + resp.Status)
	}

	payment := PaymentResponse{}

	err = json.NewDecoder(resp.Body).Decode(&payment)
	if err != nil {
		return PaymentResponse{}, err
	}

	return payment, nil
}
