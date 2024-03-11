package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	// simply respond with success code and success message
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
