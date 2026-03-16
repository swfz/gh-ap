package main

import (
	"encoding/json"
	"fmt"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"
	"log"
)

type NamedDateValue struct {
	Date githubv4.Date `json:"date,omitempty"`
}

func main() {
	// sample project
	projectId := "PVT_kwHOAD6fZs4AHaJ0"
	// contentId := "MDU6SXNzdWU5NDEyMDEyNDM="
	itemId := "PVTI_lAHOAD6fZs4AHaJ0zgDB2To"
	// Date Month
	fieldId := "PVTF_lAHOAD6fZs4AHaJ0zgERRzM"

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {date: $value}})"`
	}

	b := []byte(`{"date":"2022-09-01T00:00:00Z"}`)
	var v NamedDateValue
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId":    graphql.ID(itemId),
		"fieldId":   graphql.ID(fieldId),
		"value":     v.Date,
	}

	fmt.Println(variables)

	err = gqlclient.Mutate("Assign", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mutation)
}
