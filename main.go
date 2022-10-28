package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcool/githubv4"
	"log"
	"strconv"
	"time"
)

type Content struct {
	Id     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Option struct {
	Id   string
	Name string
}

type ProjectField struct {
	Id       string
	Name     string
	DataType string
	Options  []Option
}
type NamedDateValue struct {
	Date githubv4.Date `json:"date,omitempty"`
}

func updateDateProjectField(gqlclient api.GQLClient, projectId string, itemId string, fieldId string, fieldValue string) {
	b := []byte(`{"date":"` + fieldValue + `T00:00:00Z"}`)
	var v NamedDateValue
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {date: $value}})"`
	}
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     v.Date,
	}
	fmt.Println(variables)

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation)
}

func updateIterationProjectField(gqlclient api.GQLClient, projectId string, itemId string, fieldId string, fieldValue string) {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {iterationId: $value}})"`
	}
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     graphql.String(fieldValue),
	}
	fmt.Println(variables)

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation)
}

func updateSingleSelectProjectField(gqlclient api.GQLClient, projectId string, itemId string, fieldId string, fieldValue string) {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {singleSelectOptionId: $value}})"`
	}
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     graphql.String(fieldValue),
	}
	fmt.Println(variables)

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation)
}

func updateNumberProjectField(gqlclient api.GQLClient, projectId string, itemId string, fieldId string, fieldValue float64) {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {number: $value}})"`
	}
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     graphql.Float(fieldValue),
	}
	fmt.Println(variables)

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation)
}

func updateTextProjectField(gqlclient api.GQLClient, projectId string, itemId string, fieldId string, fieldValue string) {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {text: $value}})"`
	}
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     graphql.String(fieldValue),
	}
	fmt.Println(variables)

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation)
}

func currentPullRequestToProject(gqlclient api.GQLClient, projectId string) (itemId string) {
	args := []string{"pr", "view", "--json", "id,number,title"}
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(stdOut.String())
	var currentPR Content
	if err := json.Unmarshal(stdOut.Bytes(), &currentPR); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", currentPR)

	var addProjectMutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				Id string
			}
		} `graphql:"addProjectV2ItemById(input: {contentId: $contentId projectId: $projectId})"`
	}

	addProjectVariables := map[string]interface{}{
		"contentId": graphql.ID(currentPR.Id),
		"projectId": graphql.ID(projectId),
	}
	err = gqlclient.Mutate("Assign", &addProjectMutation, addProjectVariables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", addProjectMutation)
	return addProjectMutation.AddProjectV2ItemById.Item.Id
}

func queryProjects(gqlclient api.GQLClient, login string) (projects []struct {
	Title string
	Id    string
}) {
	var query struct {
		User struct {
			ProjectsV2 struct {
				Nodes []struct {
					Title string
					Id    string
				}
			} `graphql:"projectsV2(first: $projects)"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":    graphql.String(login),
		"projects": graphql.Int(10),
	}

	err := gqlclient.Query("ProjectsV2", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.User.ProjectsV2.Nodes
}

func queryProjectFieldTypes(gqlclient api.GQLClient, projectId string) (fieldTypes []struct {
	Id       string
	Name     string
	DataType string
}) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ProjectV2FieldCommon struct {
							Id       string
							Name     string
							DataType string
						} `graphql:"... on ProjectV2FieldCommon"`
					} `graphql:"nodes"`
				} `graphql:"fields(first: $number)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"number":    graphql.Int(20),
	}

	err := gqlclient.Query("FieldTypes", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	nodes := len(query.Node.ProjectV2.Fields.Nodes)
	fieldTypes = make([]struct {
		Id       string
		Name     string
		DataType string
	}, nodes)

	for i, node := range query.Node.ProjectV2.Fields.Nodes {
		fieldTypes[i] = node.ProjectV2FieldCommon
	}

	return fieldTypes
}

func getProjectFieldOptions(gqlclient api.GQLClient, projectId string) (fields []ProjectField) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ProjectV2IterationField struct {
							Id            string
							Name          string
							Configuration struct {
								Iterations []struct {
									StartDate string
									Id        string
								}
							}
						} `graphql:"... on ProjectV2IterationField"`
						ProjectV2SingleSelectField struct {
							Id      string
							Name    string
							Options []struct {
								Id   string
								Name string
							}
						} `graphql:"... on ProjectV2SingleSelectField"`
					} `graphql:"nodes"`
				} `graphql:"fields(first: $number)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"number":    graphql.Int(20),
	}

	err := gqlclient.Query("Fields", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	var fieldOptions []ProjectField
	for _, node := range query.Node.ProjectV2.Fields.Nodes {
		if node.ProjectV2SingleSelectField.Id != "" {
			if len(node.ProjectV2SingleSelectField.Options) > 0 {
				var options []Option
				for _, opt := range node.ProjectV2SingleSelectField.Options {
					option := Option{
						Id:   opt.Id,
						Name: opt.Name,
					}
					options = append(options, option)
				}

				field := ProjectField{
					Id:      node.ProjectV2SingleSelectField.Id,
					Name:    node.ProjectV2SingleSelectField.Name,
					Options: options,
				}
				fieldOptions = append(fieldOptions, field)
			}
		}
		if node.ProjectV2SingleSelectField.Id != "" {
			if len(node.ProjectV2IterationField.Configuration.Iterations) > 0 {
				iterations := node.ProjectV2IterationField.Configuration.Iterations

				var iterationOptions []Option
				for _, itr := range iterations {
					opt := Option{
						Id:   itr.Id,
						Name: itr.StartDate,
					}
					iterationOptions = append(iterationOptions, opt)
				}

				field := ProjectField{
					Id:      node.ProjectV2SingleSelectField.Id,
					Name:    node.ProjectV2IterationField.Name,
					Options: iterationOptions,
				}
				fieldOptions = append(fieldOptions, field)
			}
		}
	}

	return fieldOptions
}

func getProjectFields(gqlclient api.GQLClient, projectId string) (fields []ProjectField) {
	fieldOptions := getProjectFieldOptions(gqlclient, projectId)
	fieldTypes := queryProjectFieldTypes(gqlclient, projectId)

	for _, fieldType := range fieldTypes {
		field := ProjectField{
			Id:       fieldType.Id,
			Name:     fieldType.Name,
			DataType: fieldType.DataType,
		}
		if fieldType.DataType == "ITERATION" || fieldType.DataType == "SINGLE_SELECT" {
			for _, options := range fieldOptions {
				if options.Id == fieldType.Id {
					field.Options = options.Options
					break
				}
			}
		} else {
			field.Options = []Option{}
		}
		fields = append(fields, field)
	}

	return fields
}

//fmt.Printf("%+v\n", query)
//fmt.Printf("%#v\n", query)
func main() {
	restClient, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	response := struct{ Login string }{}
	err = restClient.Get("user", &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	projects := queryProjects(gqlclient, response.Login)
	projectSize := len(projects)
	projectIds := make([]string, projectSize)
	projectNames := make([]string, projectSize)
	for i, node := range projects {
		fmt.Println(i, node)
		projectIds[i] = node.Id
		projectNames[i] = node.Title
	}

	var selectedProjectId string
	q := &survey.Select{
		Message: "Choose a Project",
		Options: projectIds,
		Description: func(value string, index int) string {
			return projectNames[index]
		},
	}
	survey.AskOne(q, &selectedProjectId)
	fmt.Println("Selected Project ID", selectedProjectId)

	fields := getProjectFields(gqlclient, selectedProjectId)
	fmt.Printf("%+v\n", fields)

	itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}

	var selectedType string
	typeQuestion := &survey.Select{
		Message: "Choose a Item Type",
		Options: itemTypes,
		Default: itemTypes[0],
	}
	survey.AskOne(typeQuestion, &selectedType)

	if selectedType == "Current PullRequest" {
		itemId := currentPullRequestToProject(gqlclient, selectedProjectId)
		fmt.Println(itemId)

		for _, field := range fields {
			if field.DataType == "TEXT" {
				input := ""
				prompt := &survey.Input{
					Message: field.Name,
				}
				survey.AskOne(prompt, &input)
				if input != "" {
					updateTextProjectField(gqlclient, selectedProjectId, itemId, field.Id, input)
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
					updateDateProjectField(gqlclient, selectedProjectId, itemId, field.Id, answers[field.Name].(string))
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
				updateNumberProjectField(gqlclient, selectedProjectId, itemId, field.Id, f)
			}
			if field.DataType == "SINGLE_SELECT" || field.DataType == "ITERATION" {
				var selectedOptionId string
				message := "Choose a " + field.Name
				fieldOptionSize := len(field.Options)
				optionIds := make([]string, fieldOptionSize)
				optionNames := make([]string, fieldOptionSize)
				for i, opt := range field.Options {
					optionIds[i] = opt.Id
					optionNames[i] = opt.Name
				}
				q := &survey.Select{
					Message: message,
					Options: optionIds,
					Description: func(value string, index int) string {
						return optionNames[index]
					},
				}
				survey.AskOne(q, &selectedOptionId)
				if field.DataType == "ITERATION" {
					updateIterationProjectField(gqlclient, selectedProjectId, itemId, field.Id, selectedOptionId)
				} else {
					updateSingleSelectProjectField(gqlclient, selectedProjectId, itemId, field.Id, selectedOptionId)
				}
			}
		}
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
		number, _ := strconv.Atoi(answers["number"].(string))
		fmt.Println(number)
		var subCommand string
		if selectedType == "Issue" {
			subCommand = "issue"
		} else {
			subCommand = "pr"
		}
		args := []string{subCommand, "view", "--json", "id,number,title"}
		stdOut, _, err := gh.Exec(args...)
		if err != nil {
			fmt.Println("Error: not found " + selectedType)
			return
		}
		var content Content
		if err := json.Unmarshal(stdOut.Bytes(), &content); err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", content)
	}
}
