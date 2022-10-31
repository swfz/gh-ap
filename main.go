package main

import (
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"log"
)

func getProjects(gqlclient api.GQLClient) []Project {
	restClient, err := gh.RESTClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	response := struct{ Login string }{}
	err = restClient.Get("user", &response)
	if err != nil {
		log.Fatal(err)
	}
	var projects []Project

	userProjects := queryUserProjects(gqlclient, response.Login)
	for _, p := range userProjects {
		project := struct {
			Title string
			Id    string
			Type  string
		}{
			Title: p.Title,
			Id:    p.Id,
			Type:  "UserProject",
		}
		projects = append(projects, project)
	}

	repository := ghRepository()

	if repository.IsInOrganization {
		organizationProjects := queryOrganizationProjects(gqlclient, repository.Owner.Login)
		for _, p := range organizationProjects {
			project := struct {
				Title string
				Id    string
				Type  string
			}{
				Title: p.Title,
				Id:    p.Id,
				Type:  "OrganizationProject",
			}
			projects = append(projects, project)
		}
	}

	return projects
}

func main() {
	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	projects := getProjects(gqlclient)
	projectId := askOneProjectId(projects)
	fields := getProjectFields(gqlclient, projectId)

	itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}
	selectedType := askOneContentType(itemTypes)

	var itemId string
	if selectedType == "Current PullRequest" {
		currentPR := ghCurrentPullRequest()
		itemId = addProject(gqlclient, projectId, currentPR.Id)
	} else {
		number := askContentNumber(selectedType)
		content := ghContent(selectedType, number)
		itemId = addProject(gqlclient, projectId, content.Id)
	}

	for _, field := range fields {
		if field.DataType == "TEXT" {
			input := askTextFieldValue(field.Name)
			if input != "" {
				updateTextProjectField(gqlclient, projectId, itemId, field.Id, input)
			}
		}
		if field.DataType == "DATE" {
			dateInput := askDateFieldValue(field.Name)

			if dateInput != "" {
				updateDateProjectField(gqlclient, projectId, itemId, field.Id, dateInput)
			}
		}
		if field.DataType == "NUMBER" {
			f := askNumberFieldValue(field.Name)
			updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
		}
		if field.DataType == "SINGLE_SELECT" || field.DataType == "ITERATION" {
			selected := askOneSelectFieldValue(field.Name, field.Options)

			if selected == "Skip" {
				continue
			}
			if field.DataType == "ITERATION" {
				updateIterationProjectField(gqlclient, projectId, itemId, field.Id, selected)
			} else {
				updateSingleSelectProjectField(gqlclient, projectId, itemId, field.Id, selected)
			}
		}
	}
}
