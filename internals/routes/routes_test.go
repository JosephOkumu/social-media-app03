package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeStatic_Success(t *testing.T) {
	err := os.Chdir("../..")
	if err != nil {
		t.Fatalf("Could not change directory: %v", err)
	}
	// Ensure we change back to the original directory after the test
	defer func() {
		err := os.Chdir("internals/routes")
		if err != nil {
			t.Fatalf("Could not change back to original directory: %v", err)
		}
	}()
	// Create a response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/static/styles.css", nil)
	// Call the handler function
	serveStatic(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	responseBody := w.Body.Bytes()
	// Read the expected content from the file
	expectedContent, err := os.ReadFile("static/styles.css")
	if err != nil {
		t.Fatalf("Failed to read expected content from file: %v", err)
	}
	if !bytes.Equal(responseBody, expectedContent) {
		t.Errorf("Expected response body to match the image file content")
	}
}
