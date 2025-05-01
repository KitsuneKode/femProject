package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/kitsunekode/femProject/internal/store"
	"github.com/kitsunekode/femProject/internal/utils"
)

type UserHandler struct {
	userStore store.UserStore
	logger    *log.Logger
}

type registerNewUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore,
		logger,
	}
}

func (uh *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req registerNewUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		uh.logger.Printf("ERROR: decodingCreateUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request payload"})
		return
	}

	err = uh.validatesRegisterRequest(&req)
	if err != nil {
		uh.logger.Printf("ERROR: decodingCreateUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	user := &store.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	err = user.PasswordHash.Set(req.Password)
	if err != nil {
		uh.logger.Printf("ERROR: hashing Passowrd: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	err = uh.userStore.CreateUser(user)
	if err != nil {
		uh.logger.Printf("ERROR: createUser: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to create User"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": user})
}

func (uh *UserHandler) validatesRegisterRequest(req *registerNewUserRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}

	if len(req.Username) > 50 {
		return errors.New("username cannot be more than 50 characters")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (uh *UserHandler) HandleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		uh.logger.Printf("ERROR: readUsernameParam: invalid username parameter")
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request sent"})
		return
	}

	user, err := uh.userStore.GetUserByUsername(username)
	if err != nil {

		uh.logger.Printf("ERROR: getUserByUsername: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": user})
}

func (uh *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var user store.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		uh.logger.Printf("ERROR: updateUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request payload"})
		return
	}

	err = uh.userStore.UpdateUser(&user)
	if err == sql.ErrNoRows {
		uh.logger.Printf("ERROR: updateUser: %v", err)

		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "user doesnot exists"})
		return
	}

	if err != nil {
		uh.logger.Printf("ERROR: updateWork: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return

	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": user})
}
