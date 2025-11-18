package config

import (
	"net/http"
	"encoding/json"
	"log"
)

func RespondwithError(w http.ResponseWriter, code int, message string, err error){
	if err != nil && code >= 500 {
		log.Printf("Server error (%d): %v", code, err)
	} else if (err != nil &&  code <  500){
		log.Printf("Server error (%d): %v", code, err)
	}

	type errorResponse struct{
		Error string `json:"error"`
	}
	RespondwithJson(w, code, errorResponse{Error: message})
}


func RespondwithJson(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data, err := json.Marshal(payload)
	if err != nil{
		log.Printf("Error marshaling JSON response: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)

	// w.WriteHeader(code)
	// json.NewEncoder(w).Encode(payload)
}
