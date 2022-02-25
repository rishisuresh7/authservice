package handler

import (
	"net/http"

	"authservice/response"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	response.Success{Success: "I'm alive"}.Send(w)
}
