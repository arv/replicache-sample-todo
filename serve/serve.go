package serve

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/handlers/todo"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

const (
	awsAccessKeyId     = "REPLICANT_AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "REPLICANT_AWS_SECRET_ACCESS_KEY"
	awsRegion          = "us-west-2"
	schemaVersion      = 2
)

// Handler implements the Zeit Now entrypoint for our server.
func Handler(w http.ResponseWriter, r *http.Request) {
	name, err := dbName()
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}
	db := db.New()

	userID := authenticate(w, r)
	if userID == 0 {
		return
	}

	err = schema.Create(db, name)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}

	db.Use(name)

	switch r.URL.Path {
	case "/serve/todo-create":
		todo.Handle(w, r, db, userID)
	default:
		httperr.ClientError(w, fmt.Sprintf("Unknown path: %s", r.URL.Path))
	}
}

func authenticate(w http.ResponseWriter, r *http.Request) (userID int) {
	s := r.Header.Get("Authorization")
	if s == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Authentication required"))
		return 0
	}
	userID, err := strconv.Atoi(s)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Authorization header"))
		return 0
	}
	return userID
}

func dbName() (string, error) {
	n := "REPLICANT_SAMPLE_TODO_ENV"
	env := os.Getenv(n)
	if env == "" {
		return "", fmt.Errorf("Required environment variable %s not found", n)
	} else {
		return fmt.Sprintf("replicache_sample_todo__%s", env), nil
	}
}
