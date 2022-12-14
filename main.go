package main

import (
	"flag"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"log"
	"strconv"
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
	var options struct {
		issueNo int
		prNo    int
	}
	flag.IntVar(&options.issueNo, "issue", 0, "Issue Number")
	flag.IntVar(&options.prNo, "pr", 0, "PullRequest Number")
	flag.Parse()

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	projects := getProjects(gqlclient)
	projectId := askOneProjectId(projects)
	fields := getProjectFields(gqlclient, projectId)

	var itemId string
	var content Content

	if options.issueNo != 0 || options.prNo != 0 {
		if options.issueNo != 0 {
			content = ghContent("Issue", strconv.Itoa(options.issueNo))
		} else {
			content = ghContent("PullRequest", strconv.Itoa(options.prNo))
		}
	} else {
		itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}
		selectedType := askOneContentType(itemTypes)

		if selectedType == "Current PullRequest" {
			content = ghCurrentPullRequest()
		} else {
			contentList := ghContentList(selectedType)
			number := askContentNumber(selectedType, contentList)
			content = ghContent(selectedType, number)
		}
	}
	itemId = addProject(gqlclient, projectId, content.Id)

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
