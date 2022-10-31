package main

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"log"
	"strconv"
	"strings"
	"time"
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

	projectIds := make([]string, len(projects))
	for i, node := range projects {
		projectIds[i] = node.Id
		if node.Id == "" {
			log.Print(`[Warning] This extension requires permission for the "project" scope. You may not currently have permission to retrieve project information, please check`)
		}
	}

	qs := []*survey.Question{
		{
			Name: "ProjectId",
			Prompt: &survey.Select{
				Message: "Choose a Project",
				Options: projectIds,
				Description: func(value string, index int) string {
					return projects[index].Title + " (" + projects[index].Type + ")"
				},
				Filter: func(filterValue string, optValue string, optIndex int) bool {
					return strings.Contains(projects[optIndex].Title, filterValue)
				},
			},
		},
	}
	answers := struct{ ProjectId string }{}
	err = survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	projectId := answers.ProjectId
	fields := getProjectFields(gqlclient, projectId)

	itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}

	var selectedType string
	typeQuestion := &survey.Select{
		Message: "Choose a Item Type",
		Options: itemTypes,
		Default: itemTypes[0],
	}
	survey.AskOne(typeQuestion, &selectedType)

	var itemId string
	if selectedType == "Current PullRequest" {
		currentPR := ghCurrentPullRequest()
		itemId = addProject(gqlclient, projectId, currentPR.Id)
	} else {
		name := selectedType + " Number"
		qs := []*survey.Question{
			{
				Name:   "number",
				Prompt: &survey.Input{Message: name},
				Validate: func(v interface{}) error {
					strValue := v.(string)
					_, err := strconv.Atoi(strValue)
					if err != nil {
						return errors.New("Value is Int")
					}
					return nil
				},
			},
		}
		answers := map[string]interface{}{}
		err := survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		content := ghContent(selectedType, answers["number"].(string))
		itemId = addProject(gqlclient, projectId, content.Id)
	}
	for _, field := range fields {
		if field.DataType == "TEXT" {
			input := ""
			prompt := &survey.Input{
				Message: field.Name,
			}
			survey.AskOne(prompt, &input)
			if input != "" {
				updateTextProjectField(gqlclient, projectId, itemId, field.Id, input)
			}
		}
		if field.DataType == "DATE" {
			qs := []*survey.Question{
				{
					Name:   field.Name,
					Prompt: &survey.Input{Message: field.Name},
					Validate: func(v interface{}) error {
						strValue := v.(string)
						// Allow Zero Value
						if strValue == "" {
							return nil
						}
						_, err := time.Parse("2006-01-02", strValue)
						if err != nil {
							return errors.New("Please format it like this '2006-01-02'")
						}
						return nil
					},
				},
			}
			answers := map[string]interface{}{}
			err := survey.Ask(qs, &answers)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			if answers[field.Name] != "" {
				updateDateProjectField(gqlclient, projectId, itemId, field.Id, answers[field.Name].(string))
			}
		}
		if field.DataType == "NUMBER" {
			qs := []*survey.Question{
				{
					Name:   field.Name,
					Prompt: &survey.Input{Message: field.Name},
					Validate: func(v interface{}) error {
						strValue := v.(string)
						_, err := strconv.ParseFloat(strValue, 64)
						if err != nil {
							return errors.New("Value is Int or Float")
						}
						return nil
					},
				},
			}
			answers := map[string]interface{}{}
			err := survey.Ask(qs, &answers)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			f, _ := strconv.ParseFloat(answers[field.Name].(string), 64)
			updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
		}
		if field.DataType == "SINGLE_SELECT" || field.DataType == "ITERATION" {
			fieldOptionSize := len(field.Options)
			optionIds := make([]string, fieldOptionSize)
			for i, opt := range field.Options {
				optionIds[i] = opt.Id
			}
			qs := []*survey.Question{
				{
					Name: field.Name,
					Prompt: &survey.Select{
						Message: field.Name,
						Options: optionIds,
						Description: func(value string, index int) string {
							return field.Options[index].Name
						},
						Filter: func(filterValue string, optValue string, optIndex int) bool {
							return strings.Contains(field.Options[optIndex].Name, filterValue)
						},
					},
				},
			}
			answers := map[string]string{}
			err = survey.Ask(qs, &answers)
			if err != nil {
				log.Fatal(err.Error())
			}
			if answers[field.Name] == "Skip" {
				continue
			}
			if field.DataType == "ITERATION" {
				updateIterationProjectField(gqlclient, projectId, itemId, field.Id, answers[field.Name])
			} else {
				updateSingleSelectProjectField(gqlclient, projectId, itemId, field.Id, answers[field.Name])
			}
		}
	}
}
