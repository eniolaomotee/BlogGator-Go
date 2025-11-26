package api

import (
	"encoding/json"
	"net/http"
)

func respondWithJson(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int , message string){
	respondWithJson(w,code, ErrorResponse{Error: message})
}