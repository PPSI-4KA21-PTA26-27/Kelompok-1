package response

import (
	"encoding/json"
	"log"
	"net/http"
)

func write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("response: gagal menulis body: %v", err)
	}
}

// JSON mengirim respons sukses dengan format {"data": ...}.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, map[string]any{"data": data})
}

// Error mengirim respons gagal dengan format {"error": "..."}.
func Error(w http.ResponseWriter, status int, message string) {
	write(w, status, map[string]string{"error": message})
}

// ValidationError mengirim error per field dengan format
// {"error": "...", "fields": {"email": "..."}} supaya frontend bisa menandai input yang salah.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	write(w, http.StatusUnprocessableEntity, map[string]any{
		"error":  "data yang dikirim tidak valid",
		"fields": fields,
	})
}
