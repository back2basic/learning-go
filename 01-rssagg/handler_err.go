package main

import "net/http"

func handlerERror(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusBadRequest, "Bad Request")
}
