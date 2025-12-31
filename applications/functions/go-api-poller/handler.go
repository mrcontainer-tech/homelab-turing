package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// APIResponse represents a typical API response structure
type APIResponse struct {
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// PollRequest represents the input to the poller function
type PollRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// PollResult represents the output of the poller function
type PollResult struct {
	Success      bool                   `json:"success"`
	StatusCode   int                    `json:"status_code"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Error        string                 `json:"error,omitempty"`
	PollTime     time.Time              `json:"poll_time"`
	Duration     string                 `json:"duration"`
}

func main() {
	http.HandleFunc("/", handleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Health check for GET requests
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
		return
	}

	// Handle POST requests
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Only POST requests are supported")
		return
	}

	var input PollRequest

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// If no body provided, use default from environment
	if len(body) == 0 {
		defaultURL := os.Getenv("API_ENDPOINT")
		if defaultURL == "" {
			defaultURL = "https://api.github.com/zen"
		}
		input = PollRequest{
			URL:    defaultURL,
			Method: "GET",
		}
	} else {
		if err := json.Unmarshal(body, &input); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON input")
			return
		}
	}

	// Set default method
	if input.Method == "" {
		input.Method = "GET"
	}

	// Poll the API
	result := pollAPI(input)

	// Return result
	w.Header().Set("Content-Type", "application/json")
	if result.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(result)
}

func pollAPI(req PollRequest) PollResult {
	startTime := time.Now()
	result := PollResult{
		PollTime: startTime,
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create request
	httpReq, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(startTime).String()
		return result
	}

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set User-Agent
	httpReq.Header.Set("User-Agent", "OpenFaaS-API-Poller/1.0")

	// Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(startTime).String()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Duration = time.Since(startTime).String()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}

	// Try to parse as JSON
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		// If not JSON, store as string
		result.ResponseBody = map[string]interface{}{
			"raw": string(respBody),
		}
	} else {
		result.ResponseBody = jsonResponse
	}

	// Check if successful
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result.Success {
		result.Error = fmt.Sprintf("API returned status code: %d", resp.StatusCode)
	}

	// Here you could add logic to:
	// - Store results in a database
	// - Trigger alerts if API is down
	// - Send notifications via Slack/Discord
	// - Update metrics in Prometheus

	return result
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
