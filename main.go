package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
)

type PR struct {
	Id     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Option struct {
	Id   string
	Name string
}

type SelectField struct {
	Id      string
	Name    string
	Options []Option
}

func currentPullRequestToProject(gqlclient api.GQLClient, projectId string) (itemId string) {
	args := []string{"pr", "view", "--json", "id,number,title"}
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(stdOut.String())
	var currentPR PR
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

func getProjectFieldOptions(gqlclient api.GQLClient, projectId string) (fields []SelectField) {
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

	var fieldOptions []SelectField
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

				field := SelectField{
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

				field := SelectField{
					Id:      node.ProjectV2SingleSelectField.Id,
					Name:    node.ProjectV2IterationField.Name,
					Options: iterationOptions,
				}
				fieldOptions = append(fieldOptions, field)
			}
		}
	}

	//fmt.Println("---------------------------------")
	//fmt.Printf("%+v\n", fieldOptions)
	//fmt.Println("---------------------------------")

	return fieldOptions
}

func getProjectFields(gqlclient api.GQLClient, projectId string) (fieldTypes []struct {
	Id       string
	Name     string
	DataType string
	Options  []struct {
		Id   string
		Name string
	}
}) {

	return nil
}

//func getProjectFields(gqlclient api.GQLClient, projectId string) (projects []struct {
//	Title string
//	Id    string
//}) {
//	var query struct {
//	}
//}

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

	fields := getProjectFieldOptions(gqlclient, selectedProjectId)
	fmt.Printf("%+v\n", fields)

	//fields := getProjectFields(gqlclient, selectedProjectId)
	//fmt.Printf("%+v\n", fields)

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
	}
}
