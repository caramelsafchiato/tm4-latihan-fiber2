package middleware

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber2/helper"
)

// RequireAuth memeriksa access token pada header Authorization.
// Bila tokennya sah, identitas pemakai disimpan di Locals agar dapat
// dibaca service tanpa memeriksa ulang.
func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			// WWW-Authenticate adalah header baku yang menyertai 401.
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)

			// Membedakan "kedaluwarsa" dari "tidak valid" aman dilakukan:
			// client memang perlu tahu kapan harus memanggil /auth/refresh.
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Fail(c, fiber.StatusUnauthorized, "access token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}

	return token, nil
}

// LoginRateLimiter membatasi jumlah percobaan login dari satu alamat IP.
// Tanpa pembatasan ini, penyerang dapat mencoba ribuan password per menit
// (serangan brute force) tanpa hambatan apa pun.
func LoginRateLimiter() fiber.Handler {
	type clientLimit struct {
		count   int
		resetAt time.Time
	}
 
	const maxAttempts = 5
	var mu sync.Mutex
	clients := make(map[string]clientLimit)

	return func(c *fiber.Ctx) error {
		now := time.Now()
		ip := c.IP()

		mu.Lock()
		for key, limit := range clients {
			if !now.Before(limit.resetAt) {
				delete(clients, key)
			}
		}

		limit := clients[ip]
		if limit.count >= maxAttempts {
			mu.Unlock()
			c.Set("Retry-After", "60")
			return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi dalam satu menit")
		}
		if limit.count == 0 {
			limit.resetAt = now.Add(time.Minute)
		}
		limit.count++
		clients[ip] = limit
		mu.Unlock()

		return c.Next()
	}
}
