package main

import (
	"encoding/json"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
)

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

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
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

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
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

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
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

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
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

	err := gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
}

func addProject(gqlclient api.GQLClient, projectId string, contentId string) (itemId string) {
	var addProjectMutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				Id string
			}
		} `graphql:"addProjectV2ItemById(input: {contentId: $contentId projectId: $projectId})"`
	}

	addProjectVariables := map[string]interface{}{
		"contentId": graphql.ID(contentId),
		"projectId": graphql.ID(projectId),
	}
	err := gqlclient.Mutate("Assign", &addProjectMutation, addProjectVariables)
	if err != nil {
		log.Fatal(err)
	}

	return addProjectMutation.AddProjectV2ItemById.Item.Id
}
