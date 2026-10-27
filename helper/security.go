package helper

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
)

// passwordIterations mengatur biaya hashing password.
const passwordIterations = 310000

// dummyHash dipakai ketika username tidak ditemukan, agar waktu tanggap
// login tetap mirip dengan kasus password salah. Tanpa ini, selisih waktu
// respons membocorkan username mana yang terdaftar (timing attack).
const dummySalt = "security-dummy-salt"

// HashPassword mengubah password menjadi hash yang tidak dapat dikembalikan.
// bcrypt menyisipkan salt acak ke dalam hasilnya, sehingga dua user dengan
// password sama tetap menghasilkan hash yang berbeda.
func HashPassword(plain string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := derivePassword([]byte(plain), salt, passwordIterations)
	return "$pbkdf2-sha256$" + strconv.Itoa(passwordIterations) + "$" +
		base64.RawURLEncoding.EncodeToString(salt) + "$" +
		base64.RawURLEncoding.EncodeToString(hash), nil
}

// VerifyPassword membandingkan password dengan hash-nya.
func VerifyPassword(hash, plain string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[1] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[2])
	if err != nil || iterations < 1 || iterations > 2000000 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil || len(salt) == 0 {
		return false
	}
	expected, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil || len(expected) != sha256.Size {
		return false
	}
	actual := derivePassword([]byte(plain), salt, iterations)
	return hmac.Equal(actual, expected)
}

// VerifyDummyPassword sengaja membuang waktu seperti VerifyPassword,
// dipakai ketika username tidak ditemukan.
func VerifyDummyPassword(plain string) {
	_ = derivePassword([]byte(plain), []byte(dummySalt), passwordIterations)
}

// derivePassword implements PBKDF2-HMAC-SHA256 for one SHA-256-sized block.
func derivePassword(password, salt []byte, iterations int) []byte {
	block := append(append([]byte(nil), salt...), 0, 0, 0, 1)
	mac := hmac.New(sha256.New, password)
	_, _ = mac.Write(block)
	u := mac.Sum(nil)
	result := append([]byte(nil), u...)
	for i := 1; i < iterations; i++ {
		mac.Reset()
		_, _ = mac.Write(u)
		u = mac.Sum(nil)
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}

// RandomToken menghasilkan string acak yang aman secara kriptografis.
// math/rand TIDAK boleh dipakai untuk keperluan ini karena hasilnya
// dapat diperkirakan bila seed-nya diketahui.
func RandomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SHA256Hex dipakai untuk menyimpan refresh token dalam bentuk hash.
// bcrypt tidak dipakai di sini karena token sudah acak dan panjang,
// sehingga tidak perlu diperlambat seperti password buatan manusia.
func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
