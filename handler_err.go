package main

import "net/http"

func handlerError(w http.ResponseWriter, r *http.Request) {
	// simply respond with an error message and error code
	respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}
