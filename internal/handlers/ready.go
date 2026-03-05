package handlers

import (
	"net/http"

	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

func HelloReady(w http.ResponseWriter, r *http.Request) { helpers.RespondWithJson(w, 200, "hello") }
func ErrorReady(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithError(w, 200, "this is an error test")
}
