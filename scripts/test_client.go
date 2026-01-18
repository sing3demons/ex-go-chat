package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func main() {
	const (
		totalRequests   = 10000
		concurrentUsers = 100
	)

	requestsPerUser := totalRequests / concurrentUsers

	var wg sync.WaitGroup
	var successCount, errorCount int
	var mu sync.Mutex

	startTime := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			for j := 0; j < requestsPerUser; j++ {
				reqBody := `{"identifier": "chattest1", "password": "Test1234567!"}`

				req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/auth/login", strings.NewReader(reqBody))
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				// Read and discard body to allow connection reuse
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				mu.Lock()
				if resp.StatusCode == http.StatusOK {
					successCount++
				} else {
					errorCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("=== Load Test Results ===\n")
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Concurrent Users: %d\n", concurrentUsers)
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Requests/sec: %.2f\n", float64(totalRequests)/elapsed.Seconds())
}
