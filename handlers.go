package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func handleProjects(roster *Roster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := roster.ListAllProjects()
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(projects); err != nil {
			handleError(w, err)
			return
		}
	}
}

func handleRoster(roster *Roster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/roster/")

		includesStr := r.URL.Query().Get("include")
		var includes map[string]struct{}
		includes = make(map[string]struct{})
		for _, include := range strings.Split(includesStr, ",") {
			includes[include] = struct{}{}
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			handleError(w, err)
			return
		}

		includeAssignee := false
		if _, ok := includes["assignee"]; ok {
			includeAssignee = true
		}

		since, err := time.Parse(time.RFC3339, r.URL.Query().Get("since"))
		if err != nil {
			since = time.Now().AddDate(0, 0, -15)
		}

		users, err := roster.GetRoster(id, since, includeAssignee)
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			handleError(w, err)
			return
		}
	}
}

func handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		log.Fatal(err)
	}
}
