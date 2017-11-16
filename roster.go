package main

import (
	"strconv"
	"strings"
	"time"

	gitlab "github.com/xanzy/go-gitlab"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"name_with_namespace"`
}

func NewProjectFromGitlabProject(gitlabProject *gitlab.Project) *Project {
	return &Project{
		ID:                gitlabProject.ID,
		Name:              gitlabProject.Name,
		NameWithNamespace: strings.Replace(gitlabProject.NameWithNamespace, " ", "", -1),
	}
}

type User struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type Roster struct {
	client *gitlab.Client
}

func NewRoster(gitlabAccessToken, gitlabApiUrl string) *Roster {
	client := gitlab.NewClient(nil, gitlabAccessToken)
	client.SetBaseURL(gitlabApiUrl)

	return &Roster{
		client: client,
	}
}

func (r *Roster) ListAllProjects() ([]*Project, error) {
	var projects []*Project

	page := 1

	for page != 0 {
		options := gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}
		ps, response, err := r.client.Projects.ListProjects(&options)

		if err != nil {
			return nil, err
		}

		for _, p := range ps {
			projects = append(projects, NewProjectFromGitlabProject(p))
		}

		page, err = strconv.Atoi(response.Header.Get("X-Next-Page"))
		if err != nil {
			page = 0
		}
	}

	return projects, nil
}

// FIXME return nil if pid is unknown
func (r *Roster) GetRoster(pid int, since time.Time, includeAssignee bool) ([]*User, error) {
	var users []*User
	mrs, _, err := r.client.MergeRequests.ListProjectMergeRequests(pid, &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100, // should not be more than 100 opened merge requests
		},
		State: gitlab.String("opened"),
	})

	if err != nil {
		return nil, err
	}

	for _, mr := range mrs {
		users = append(users, extractUsersFromMergeRequest(mr, includeAssignee)...)
	}

	var mergeRequests []*gitlab.MergeRequest

	mergeRequests, err = r.listProjectMergeRequestsSince(pid, since, "merged")
	if err != nil {
		return nil, err
	}

	for _, mr := range mergeRequests {
		users = append(users, extractUsersFromMergeRequest(mr, includeAssignee)...)
	}

	mergeRequests, err = r.listProjectMergeRequestsSince(pid, since, "closed")
	if err != nil {
		return nil, err
	}

	for _, mr := range mergeRequests {
		users = append(users, extractUsersFromMergeRequest(mr, includeAssignee)...)
	}

	return dedupUsers(users), nil
}

func (r *Roster) listProjectMergeRequestsSince(pid int, since time.Time, state string) ([]*gitlab.MergeRequest, error) {
	var mergeRequests []*gitlab.MergeRequest

	page := 1
	date := time.Now()

	for since.Before(date) && page != 0 {
		mrs, response, err := r.client.MergeRequests.ListProjectMergeRequests(pid, &gitlab.ListProjectMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 20, // should not be more than 100 opened merge requests
			},
			State:   gitlab.String(state),
			OrderBy: gitlab.String("updated_at"),
			Sort:    gitlab.String("desc"),
		})
		if err != nil {
			return nil, err
		}

		for _, mr := range mrs {
			updatedAt, _ := time.Parse(time.RFC3339, mr.UpdatedAt)
			if since.Before(updatedAt) {
				mergeRequests = append(mergeRequests, mr)
			}
		}

		page, err = strconv.Atoi(response.Header.Get("X-Next-Page"))
		if err != nil {
			page = 0
		}

		date, _ = time.Parse(time.RFC3339, mrs[len(mrs)-1].UpdatedAt)
	}

	return mergeRequests, nil
}

func extractUsersFromMergeRequest(mr *gitlab.MergeRequest, includeAssignee bool) []*User {
	var users []*User

	users = append(users, &User{
		Name:      mr.Author.Name,
		Username:  mr.Author.Username,
		AvatarURL: mr.Author.AvatarURL,
	})

	if includeAssignee && mr.Assignee.Username != "" {
		users = append(users, &User{
			Name:      mr.Assignee.Name,
			Username:  mr.Assignee.Username,
			AvatarURL: mr.Assignee.AvatarURL,
		})
	}

	return users
}

func dedupUsers(users []*User) []*User {
	var usersMap map[string]*User
	usersMap = make(map[string]*User)

	for _, u := range users {
		usersMap[u.Username] = u
	}

	var dedupUsers []*User

	for _, u := range usersMap {
		dedupUsers = append(dedupUsers, u)
	}

	return dedupUsers
}
