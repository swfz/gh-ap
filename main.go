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
type Project struct {
	Id    string
	Title string
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

	//fmt.Printf("%+v\n", query)
	//fmt.Printf("%#v\n", query)

	projects := queryProjects(gqlclient, response.Login)
	projectSize := len(projects)
	projectIds := make([]string, projectSize)
	projectNames := make([]string, projectSize)
	for i, node := range projects {
		fmt.Println(i, node)
		projectIds[i] = node.Id
		projectNames[i] = node.Title
	}

	var selectedId string
	q := &survey.Select{
		Message: "Choose a Project",
		Options: projectIds,
		Description: func(value string, index int) string {
			return projectNames[index]
		},
	}

	survey.AskOne(q, &selectedId)
	fmt.Println("Selected Project ID", selectedId)

	itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}

	var selectedType string
	typeQuestion := &survey.Select{
		Message: "Choose a Item Type",
		Options: itemTypes,
		Default: itemTypes[0],
	}
	survey.AskOne(typeQuestion, &selectedType)

	if selectedType == "Current PullRequest" {
		itemId := currentPullRequestToProject(gqlclient, selectedId)
		fmt.Println(itemId)
	}
}
