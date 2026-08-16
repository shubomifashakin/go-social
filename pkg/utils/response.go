package utils

import (
	"encoding/json"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, code int, body any) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)

    return json.NewEncoder(w).Encode(body)
}
