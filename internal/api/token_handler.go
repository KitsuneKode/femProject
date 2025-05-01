package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kitsunekode/femProject/internal/store"
	"github.com/kitsunekode/femProject/internal/tokens"
	"github.com/kitsunekode/femProject/internal/utils"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewTokenHandler(tokenStore store.TokenStore, logger *log.Logger, userStore store.UserStore) *TokenHandler {
	return &TokenHandler{
		tokenStore,
		userStore,
		logger,
	}
}

func (th *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		th.logger.Printf("ERROR: decodeRequestToken: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid reuest payload"})
		return
	}

	user, err := th.userStore.GetUserByUsername(req.Username)
	if err != nil || user == nil {

		th.logger.Printf("ERROR: getUserByUsername: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	passwordDoMatch, err := user.PasswordHash.Matches(req.Password)
	if err != nil {

		th.logger.Printf("ERROR: checkPasswordMatch: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "interal server error"})
		return
	}

	if !passwordDoMatch {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid credentials"})
		return
	}

	token, err := th.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {

		th.logger.Printf("ERROR: createNewToken: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"auth_token": token})
}
