package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"latihan-fiber2/app/model"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

// accessClaims adalah isi access token. Selain field bawaan JWT,
// ditambahkan username dan role agar middleware tidak perlu
// menanyakannya ke database pada setiap request.
type accessClaims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	Subject   string `json:"sub"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewJWTManager(secret, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

func (m *JWTManager) AccessTTL() time.Duration { return m.accessTTL }

// GenerateAccess membuat access token berumur pendek.
func (m *JWTManager) GenerateAccess(u model.User) (string, error) {
	now := time.Now()
	claims := accessClaims{
		Username:  u.Username,
		Role:      u.Role,
		Subject:   strconv.Itoa(u.ID),
		Issuer:    m.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(m.accessTTL).Unix(),
	}
	header, err := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + m.sign(unsigned), nil
}

// Parse memeriksa tanda tangan dan masa berlaku token,
// lalu mengembalikan identitas yang dibawanya.
func (m *JWTManager) Parse(tokenString string) (model.AuthUser, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return model.AuthUser{}, ErrInvalidToken
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if json.Unmarshal(headerJSON, &header) != nil || header.Algorithm != "HS256" {
		return model.AuthUser{}, ErrInvalidToken
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	expectedSignature, err := base64.RawURLEncoding.DecodeString(m.sign(unsigned))
	if err != nil || !hmac.Equal(providedSignature, expectedSignature) {
		return model.AuthUser{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}
	var claims accessClaims
	if json.Unmarshal(payload, &claims) != nil || claims.Issuer != m.issuer || claims.ExpiresAt == 0 {
		return model.AuthUser{}, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return model.AuthUser{}, ErrExpiredToken
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}

	return model.AuthUser{
		UserID:   userID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func (m *JWTManager) sign(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
