package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "http://localhost:3000/api/v1"
const dummyPassword = "password"

func main() {
	mode := flag.String("mode", "crowd", "Mode pengujian: crowd, credits, swap, supervisor")
	flag.Parse()

	fmt.Printf("Memulai uji beban mode: %s\n", *mode)

	if *mode == "crowd" {
		runCrowdTest()
	} else {
		fmt.Println("Mode belum diimplementasikan di skrip basic ini.")
	}
}

func getDummyToken(studentID int) string {
	username := fmt.Sprintf("mhs%03d", studentID)
	loginReq := map[string]string{"username": username, "password": dummyPassword}
	reqBody, _ := json.Marshal(loginReq)

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("User %s gagal: Server mati atau URL salah (%v)\n", username, err)
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("User %s gagal login! Respons Server: %s\n", username, string(bytes.TrimSpace(bodyBytes)))
		return ""
	}

	var result struct {
		Data struct {
			AccessToken string `json:"access_token"`
			Token       string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(bodyBytes, &result)

	if result.Data.AccessToken != "" {
		return result.Data.AccessToken
	}
	return result.Data.Token
}

func runCrowdTest() {
	var wg sync.WaitGroup
	startGun := make(chan struct{})
	totalUsers := 60
	courseID := 1 

	fmt.Printf("Menyiapkan %d goroutine untuk menyerang endpoint pendaftaran...\n", totalUsers)

	for i := 1; i <= totalUsers; i++ {
		wg.Add(1)
		go func(studentID int) {
			defer wg.Done()

			token := getDummyToken(studentID)
			if token == "" {
				<-startGun 
				return
			}

			<-startGun

			req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", baseURL, courseID), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("User %d: Error %v\n", studentID, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 201 {
				fmt.Printf("User %d: BERHASIL (201)\n", studentID)
			} else {
				fmt.Printf("User %d: GAGAL (%d) - %s\n", studentID, resp.StatusCode, string(bytes.TrimSpace(body)))
			}
		}(i)
	}

	time.Sleep(3 * time.Second)
	fmt.Println("\nDOR! Semua request ditembakkan bersamaan!")
	close(startGun)
	wg.Wait()
	fmt.Println("Uji beban selesai.")
}