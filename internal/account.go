package internal

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"scratch/api"
	userManager "scratch/internal/services"
)

var _ api.ServerInterface = (*accountHandler)(nil)

type accountHandler struct {
	am  userManager.AccountManager
	log slog.Logger
}

func NewAccountHandler(am userManager.AccountManager, log slog.Logger) *accountHandler {
	return &accountHandler{am: am, log: log}
}

func (ah *accountHandler) PostRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request api.RegisterUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		ah.writeJSON(w, http.StatusBadRequest, "can not create user account")
		return
	}

	id, err := ah.am.CreateUser(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, userManager.UserExistErr):
			ah.writeJSON(w, http.StatusConflict, api.ErrorResponse{Error: "user with that email already exists"})
			return
		default:
			ah.writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: "Internal server error"})
			return
		}
	}
	responseBody := struct {
		Id int `json:"id"`
	}{
		Id: id,
	}

	ah.writeJSON(w, http.StatusCreated, responseBody)
	return
}

func (ah *accountHandler) GetUserId(w http.ResponseWriter, r *http.Request, id int) {
	//TODO implement me
	panic("implement me")
}

func (ah *accountHandler) PostLogin(w http.ResponseWriter, r *http.Request) {

	var body api.PostLoginJSONRequestBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		ah.writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "decode request error"})
		return
	}
	response, err := ah.am.Login(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, userManager.UserNotFoundErr):
			ah.writeJSON(w, http.StatusNotFound, api.ErrorResponse{Error: "user not found"})
			return
		case errors.Is(err, userManager.IncorrectPasswordErr):
			ah.writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "incorrect credentials"})
		default:
			ah.writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: "internal server error"})
			return
		}
	}
	ah.writeJSON(w, http.StatusOK, response)
}

func (ah *accountHandler) writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		ah.log.Warn("encode response: %w", err)
	}
}
