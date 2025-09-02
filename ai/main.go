package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os" // New import for working directory
	"path/filepath"
)

// PythonRequest is the structure for the request body sent to the Python service.
type PythonRequest struct {
	Prompt string `json:"prompt"`
}

// PythonResponse is the structure for the response body received from the Python service.
type PythonResponse struct {
	Response string `json:"response"`
}

// chatHandler handles requests from the web page and forwards them to the Python service.
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Set CORS headers to allow requests from the web page.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Parse the request body from the front-end.
	var requestBody struct {
		Message string `json:"message"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Prepare the request body to send to the Python service.
	pyReq := PythonRequest{
		Prompt: requestBody.Message,
	}
	pyReqBytes, _ := json.Marshal(pyReq)

	// Call the Python service.
	url := "http://localhost:8000/generate"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(pyReqBytes))
	if err != nil {
		log.Printf("Failed to call Python service: %v", err)
		http.Error(w, "Unable to connect to the large language model service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Python service returned an error status: %d", resp.StatusCode)
		http.Error(w, "Large language model service returned an error", http.StatusBadGateway)
		return
	}

	// Read the response from the Python service.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Python response: %v", err)
		http.Error(w, "Failed to read the large language model response", http.StatusInternalServerError)
		return
	}

	var pyResp PythonResponse
	err = json.Unmarshal(bodyBytes, &pyResp)
	if err != nil {
		log.Printf("Failed to parse Python response: %v", err)
		http.Error(w, "Failed to parse the large language model response", http.StatusInternalServerError)
		return
	}

	// Return the model's response to the web page.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseToClient := map[string]string{
		"response": pyResp.Response,
	}
	json.NewEncoder(w).Encode(responseToClient)
}

// homeHandler serves the index.html file from the current working directory.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current working directory.
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Join the current directory with the file name to get the full path.
	filePath := filepath.Join(dir, "chat.html")

	// Serve the index.html file.
	http.ServeFile(w, r, filePath)
}

func main() {
	// Set up the routes.
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/chat", chatHandler)

	fmt.Println("Go service started, listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
