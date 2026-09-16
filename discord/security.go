package discord

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	// ErrInvalidSignature is returned when the Ed25519 signature fails verification.
	ErrInvalidSignature = errors.New("invalid interaction ed25519 signature")
	// ErrMissingHeaders is returned when required signature headers are absent.
	ErrMissingHeaders = errors.New("missing signature or timestamp headers")
)

// VerifyInteractionRequest verifies the Ed25519 signature of an incoming Discord HTTP interaction request.
func VerifyInteractionRequest(publicKeyHex string, headers http.Header, body []byte) bool {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	sigHex := headers.Get("X-Signature-Ed25519")
	timestamp := headers.Get("X-Signature-Timestamp")
	if sigHex == "" || timestamp == "" {
		return false
	}

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}

	msg := append([]byte(timestamp), body...)
	return ed25519.Verify(pubKeyBytes, msg, sigBytes)
}

// InteractionHandler represents a callback that processes a verified Discord interaction.
type InteractionHandler func(interaction *Interaction) *InteractionResponse

// NewHTTPInteractionHandler creates an http.HandlerFunc that verifies incoming requests using the public key
// and handles Discord PING interactions automatically.
func NewHTTPInteractionHandler(publicKeyHex string, handler InteractionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if !VerifyInteractionRequest(publicKeyHex, r.Header, body) {
			http.Error(w, "Unauthorized: Invalid Signature", http.StatusUnauthorized)
			return
		}

		var interaction Interaction
		if err := json.Unmarshal(body, &interaction); err != nil {
			http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
			return
		}

		// Handle Discord PING validation automatically
		if interaction.Type == InteractionTypePing {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(InteractionResponse{
				Type: InteractionCallbackTypePong,
			})
			return
		}

		// Dispatch to user-provided interaction handler
		if handler != nil {
			resp := handler(&interaction)
			if resp != nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
