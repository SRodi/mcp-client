package netclient

import (
	"context"
	"testing"
	"time"
)

func TestClient_NewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("Expected baseURL to be 'http://localhost:8080', got '%s'", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestClient_Connect(t *testing.T) {
	client := NewClient("http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should fail since there's no server running
	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected connection to fail when no server is running")
		client.Close() // Clean up if somehow it connected
	}
}
