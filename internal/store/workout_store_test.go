package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=secretpassword dbname=postgres port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	err = Migrate(db, "../../migrations/")
	if err != nil {
		t.Fatalf("migrating test db error: %v", err)
	}

	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("truncating test db error: %v", err)
	}
	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name: "valid workout",
			workout: &Workout{
				Title:           "push day",
				Description:     "upper body day",
				DurationMinutes: 60,
				CaloriesBurned:  300,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Bench Press",
						Sets:         3,
						Reps:         IntPtr(10),
						Weight:       FloatPtr(135.5),
						// DurationSeconds: IntPtr(45),
						Notes:      " warm up properly",
						OrderIndex: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid workout",
			workout: &Workout{
				Title:           "pull day",
				Description:     "back day",
				DurationMinutes: 45,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Deadlift",
						Sets:         3,
						Reps:         IntPtr(8),
						// Weight:          FloatPtr(225.0),
						// DurationSeconds: IntPtr(60),
						Notes:      " focus on form",
						OrderIndex: 1,
					},

					{
						ExerciseName:    "Pull Up",
						Sets:            3,
						Reps:            IntPtr(8),
						DurationSeconds: IntPtr(30),
						Weight:          FloatPtr(165.0),
						Notes:           " use a band",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.CaloriesBurned, createdWorkout.CaloriesBurned)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)

			retrieved, err := store.GetWorkoutByID(int64(createdWorkout.ID))

			require.NoError(t, err)

			assert.Equal(t, createdWorkout.ID, retrieved.ID)
			assert.Equal(t, createdWorkout.Title, retrieved.Title)
			assert.Equal(t, createdWorkout.Description, retrieved.Description)
			assert.Equal(t, createdWorkout.DurationMinutes, retrieved.DurationMinutes)

			// assert.Equal(t, createdWorkout.Entries, retrieved.Entries)

			assert.Equal(t, len(tt.workout.Entries), len(retrieved.Entries))

			for i, entry := range retrieved.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, entry.ExerciseName)
				assert.Equal(t, tt.workout.Entries[i].DurationSeconds, entry.DurationSeconds)
				assert.Equal(t, tt.workout.Entries[i].Notes, entry.Notes)
				assert.Equal(t, tt.workout.Entries[i].OrderIndex, entry.OrderIndex)
				assert.Equal(t, tt.workout.Entries[i].Reps, entry.Reps)
				assert.Equal(t, tt.workout.Entries[i].Sets, entry.Sets)
			}
		})
	}
}

func IntPtr(i int) *int {
	return &i
}

func FloatPtr(i float64) *float64 {
	return &i
}
