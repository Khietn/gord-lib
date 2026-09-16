package discord

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func generateTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey, string) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}
	return pub, priv, hex.EncodeToString(pub)
}

func TestVerifyInteractionRequest(t *testing.T) {
	_, priv, pubHex := generateTestKeyPair(t)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"type":1}`)
	msg := append([]byte(timestamp), body...)
	sig := ed25519.Sign(priv, msg)
	sigHex := hex.EncodeToString(sig)

	headers := make(http.Header)
	headers.Set("X-Signature-Ed25519", sigHex)
	headers.Set("X-Signature-Timestamp", timestamp)

	// Valid signature
	if !VerifyInteractionRequest(pubHex, headers, body) {
		t.Errorf("expected valid signature verification to pass")
	}

	// Corrupted body
	if VerifyInteractionRequest(pubHex, headers, []byte(`{"type":2}`)) {
		t.Errorf("expected tampered body to fail verification")
	}

	// Corrupted timestamp
	tamperedHeaders := headers.Clone()
	tamperedHeaders.Set("X-Signature-Timestamp", "12345678")
	if VerifyInteractionRequest(pubHex, tamperedHeaders, body) {
		t.Errorf("expected tampered timestamp to fail verification")
	}

	// Missing header
	missingHeaders := make(http.Header)
	missingHeaders.Set("X-Signature-Timestamp", timestamp)
	if VerifyInteractionRequest(pubHex, missingHeaders, body) {
		t.Errorf("expected missing signature header to fail verification")
	}

	// Invalid hex public key
	if VerifyInteractionRequest("invalid-hex", headers, body) {
		t.Errorf("expected invalid hex public key to fail")
	}
}

func TestNewHTTPInteractionHandler(t *testing.T) {
	_, priv, pubHex := generateTestKeyPair(t)

	handler := NewHTTPInteractionHandler(pubHex, func(interaction *Interaction) *InteractionResponse {
		return &InteractionResponse{
			Type: InteractionCallbackTypeChannelMessageWithSource,
			Data: &InteractionCallbackData{
				Content: "Handled custom interaction",
			},
		}
	})

	signRequest := func(req *http.Request, body []byte) {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		msg := append([]byte(timestamp), body...)
		sig := ed25519.Sign(priv, msg)
		req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(sig))
		req.Header.Set("X-Signature-Timestamp", timestamp)
	}

	// 1. Method Not Allowed (GET)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/interactions", nil)
	handler(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected HTTP 405 for GET, got %d", rec.Code)
	}

	// 2. Unauthorized (invalid signature)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader([]byte(`{"type":1}`)))
	req.Header.Set("X-Signature-Ed25519", "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	req.Header.Set("X-Signature-Timestamp", "12345")
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 for bad signature, got %d", rec.Code)
	}

	// 3. Valid PING (Type: 1) -> Automatic PONG
	pingBody := []byte(`{"type":1}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(pingBody))
	signRequest(req, pingBody)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for valid ping, got %d", rec.Code)
	}

	var pongResp InteractionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &pongResp); err != nil {
		t.Fatalf("failed to decode pong response: %v", err)
	}
	if pongResp.Type != InteractionCallbackTypePong {
		t.Errorf("expected InteractionCallbackTypePong (1), got %d", pongResp.Type)
	}

	// 4. Custom Application Command Interaction (Type: 2)
	cmdBody := []byte(`{"type":2,"id":"12345"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(cmdBody))
	signRequest(req, cmdBody)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for command, got %d", rec.Code)
	}

	var cmdResp InteractionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &cmdResp); err != nil {
		t.Fatalf("failed to decode command response: %v", err)
	}
	if cmdResp.Data.Content != "Handled custom interaction" {
		t.Errorf("expected content 'Handled custom interaction', got %s", cmdResp.Data.Content)
	}
}
