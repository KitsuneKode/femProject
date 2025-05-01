package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kitsunekode/femProject/internal/middleware"
	"github.com/kitsunekode/femProject/internal/store"
	"github.com/kitsunekode/femProject/internal/utils"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *log.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore,
		logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR: readIDParam: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout ID"})
		return
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("ERROR: getWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout})
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		wh.logger.Printf("ERROR: decodingCreateWorkout: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request sent"})
		return
	}

	currentUser := middleware.GetUser(r)

	if currentUser == nil || currentUser == store.AnonnymousUser {
		wh.logger.Printf("ERROR: createWorkoutNoUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "you must be logged in"})
		return
	}

	workout.UserID = currentUser.ID

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		wh.logger.Printf("ERROR: createWorkout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to create Workout"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"workout": createdWorkout})
}

func (wh *WorkoutHandler) HandleUpdateWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR: readIDParam: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout ID"})
		return
	}
	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutID)

	if existingWorkout == nil {
		wh.logger.Printf("ERROR: updateWorkoutByID: No workout with the ID found")
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout ID"})
		return
	}

	if err != nil {
		wh.logger.Printf("ERROR: updateWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	var updateWorkoutRequest struct {
		Title           *string              `json:"title"`
		Description     *string              `json:"description"`
		DurationMinutes *int                 `json:"duration_minutes"`
		CalorieBurned   *int                 `json:"calorie_burned"`
		Entries         []store.WorkoutEntry `json:"entries"`
	}

	err = json.NewDecoder(r.Body).Decode(&updateWorkoutRequest)
	if err != nil {
		wh.logger.Printf("ERROR: updateWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	if updateWorkoutRequest.Title != nil {
		existingWorkout.Title = *updateWorkoutRequest.Title
	}
	if updateWorkoutRequest.Description != nil {
		existingWorkout.Description = *updateWorkoutRequest.Description
	}

	if updateWorkoutRequest.CalorieBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutRequest.CalorieBurned
	}

	if updateWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updateWorkoutRequest.Entries
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == store.AnonnymousUser {
		wh.logger.Printf("ERROR: updateWorkoutNoUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "you must be logged in"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			{
				wh.logger.Printf("ERROR: updateWorkoutNoWorkout: %v", err)
				utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "workout does not exists"})
				return
			}
		default:
			{
				wh.logger.Printf("ERROR: updateWorkout: %v", err)
				utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
				return
			}

		}
	}

	if workoutOwner != currentUser.ID {

		wh.logger.Printf("ERROR: updateWorkoutNotOwner: %v", err)
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "you are not authorized to update this workout"})
		return
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout)
	if err != nil {
		wh.logger.Printf("ERROR: updateWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"updatedWorkout": existingWorkout})
}

func (wh *WorkoutHandler) HandleDeleteWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Printf("ERROR:readIDParam : %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout ID"})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == store.AnonnymousUser {
		wh.logger.Printf("ERROR: updateWorkoutNoUser: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "you must be logged in"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			wh.logger.Printf("ERROR: deleteWorkoutNoWorkout: %v", err)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "workout does not exists"})
			return
		}
		wh.logger.Printf("ERROR: deleteWorkout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return

	}

	if workoutOwner != currentUser.ID {

		wh.logger.Printf("ERROR: deleteWorkoutNotOwner: %v", err)
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "you are not authorized to delete this workout"})
		return
	}

	err = wh.workoutStore.DeleteWorkoutByID(workoutID)
	if err == sql.ErrNoRows {
		wh.logger.Printf("ERROR: deleteWorkoutByID: %v", sql.ErrNoRows)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout ID"})
		return
	}

	if err != nil {
		wh.logger.Printf("ERROR: deleteWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workoutID": workoutID})
}
