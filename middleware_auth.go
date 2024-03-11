package main

import (
	"fmt"
	"net/http"

	"github.com/yuheng-liu/rssaggregator/internal/auth"
	"github.com/yuheng-liu/rssaggregator/internal/database"
)

// used to convert a custom auth handler to be in the same signature as http.HandlerFunc
// this custom handler has an additional database.User parameter
type authedHandler func(http.ResponseWriter, *http.Request, database.User)

// take in a custom handler and returns the standard http.HandlerFunc
func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	// authentication code for use with any api that needs authentication
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusForbidden, fmt.Sprintf("Auth error: %v", err))
			return
		}

		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Couldn't get user: %v", err))
			return
		}

		handler(w, r, user)
	}
}
