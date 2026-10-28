package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "http://localhost:3000/api/v1"

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

func getToken(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var loginRes LoginResponse
	_ = json.Unmarshal(respBody, &loginRes)

	if !loginRes.Success {
		return "", fmt.Errorf("gagal login user %s: %s", username, string(respBody))
	}
	return loginRes.Data.Token, nil
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("   PENGUJIAN BEBAN & CONCURRENCY (MODUL 10)")
	fmt.Println("==================================================")

	// 1. Ambil Token Mahasiswa
	token, err := getToken("alisyaaninda", "alisyaaninda123")
	if err != nil {
		fmt.Printf("Error Login: %v\n", err)
		return
	}
	fmt.Println("[OK] Login Berhasil. Menjalankan pengujian serentak...\n")

	// 2. Pengujian Concurrency Batas 20 Proyek (10 Request Serentak)
	fmt.Println("--- 1. Uji Penambahan 10 Proyek Serentak ---")
	testPortfolioProjectLimit(token)

	// 3. Pengujian Race Condition Vote Helpful (1 User, 20 Request Serentak)
	fmt.Println("\n--- 2. Uji Race Condition Vote (1 User 20 Req Serentak) ---")
	testRaceConditionHelpful(token)
}

func testPortfolioProjectLimit(token string) {
	studentID := 2 // Sesuaikan dengan ID Mahasiswa di PostgreSQL kamu
	url := fmt.Sprintf("%s/students/%d/portfolio/projects", baseURL, studentID)

	var wg sync.WaitGroup
	startGun := make(chan struct{})
	totalReq := 10

	var successCount, conflictCount, otherCount int
	var mu sync.Mutex

	for i := 1; i <= totalReq; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			// Menunggu sinyal tembak bersamaan
			<-startGun

			body, _ := json.Marshal(map[string]interface{}{
				"title":        fmt.Sprintf("Proyek LoadTest #%d", reqNum),
				"technologies": []string{"Go", "Fiber", "MongoDB"},
				"link":         "https://github.com/caramelsafchiato",
				"year":         2026,
			})

			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("[Req #%d] Error: %v\n", reqNum, err)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusCreated {
				successCount++
				fmt.Printf("[Req #%d] -> 201 Created (Proyek Berhasil Ditambah)\n", reqNum)
			} else if resp.StatusCode == http.StatusConflict {
				conflictCount++
				fmt.Printf("[Req #%d] -> 409 Conflict (Ditolak: Kuota Max 20 Terpenuhi)\n", reqNum)
			} else {
				otherCount++
				fmt.Printf("[Req #%d] -> Status HTTP %d\n", reqNum, resp.StatusCode)
			}
			mu.Unlock()
		}(i)
	}

	// Pelepasan sinyal agar 10 Goroutines jalan bersamaan
	close(startGun)
	wg.Wait()

	fmt.Printf("\n--> REKAP BUKTI LIMIT PROYEK: 201 Created = %d | 409 Conflict = %d | Lainnya = %d\n",
		successCount, conflictCount, otherCount)
}

func testRaceConditionHelpful(token string) {
	// Ganti string ini dengan ID ulasan ObjectID MongoDB yang valid di DB-mu
	reviewID := "66e01a2b3c4d5e6f7a8b9c0d" 
	url := fmt.Sprintf("%s/reviews/%s/helpful", baseURL, reviewID)

	var wg sync.WaitGroup
	startGun := make(chan struct{})
	totalReq := 20

	var successCount, conflictCount int
	var mu sync.Mutex

	for i := 1; i <= totalReq; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			<-startGun

			req, _ := http.NewRequest("POST", url, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
				fmt.Printf("[Req #%d] -> 200 OK (Vote Berhasil)\n", reqNum)
			} else if resp.StatusCode == http.StatusConflict {
				conflictCount++
				fmt.Printf("[Req #%d] -> 409 Conflict (Ditolak: Sudah Vote/Ulasan Sendiri)\n", reqNum)
			}
			mu.Unlock()
		}(i)
	}

	close(startGun)
	wg.Wait()

	fmt.Printf("\n--> REKAP BUKTI RACE CONDITION: 200 OK = %d | 409 Conflict = %d\n",
		successCount, conflictCount)
}