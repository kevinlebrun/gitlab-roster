package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	var (
		gitlabAccessToken = flag.String("gitlab-access-token", "", "Your Gitlab's application access token")
		gitlabApiUrl      = flag.String("gitlab-api-url", "", "Your Gitlab API endpoint")
		port              = flag.String("port", "8080", "The listening port of Roster")
	)
	flag.Parse()

	roster := NewRoster(*gitlabAccessToken, *gitlabApiUrl)

	http.HandleFunc("/projects", handleProjects(roster))
	http.HandleFunc("/roster/", handleRoster(roster))

	fmt.Printf("Listening on port %s...\n", *port)
	http.ListenAndServe(":"+*port, nil)
}
