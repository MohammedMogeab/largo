package handlers

import (
    "net/http"
)

type UserController struct{}

func (h *UserController) Handle(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}
