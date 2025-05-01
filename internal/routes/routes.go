package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/kitsunekode/femProject/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(app.Middleware.Authenticate)

		r.Get("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleGetWorkoutByID))
		r.Post("/workouts", app.Middleware.RequireUser(app.WorkoutHandler.HandleCreateWorkout))
		r.Put("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleUpdateWorkoutByID))
		r.Delete("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleDeleteWorkoutByID))

		r.Put("/users/{username}", app.Middleware.RequireUser(app.UserHandler.HandleUpdateUser))
	})

	r.Get("/health", app.HealthCheck)

	r.Get("/users/{username}", app.UserHandler.HandleGetUserByUsername)
	r.Post("/user", app.UserHandler.HandleCreateUser)

	r.Post("/tokens/authentication", app.TokenHandler.HandleCreateToken)

	return r
}
