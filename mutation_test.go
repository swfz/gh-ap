package main

import (
	graphql "github.com/cli/shurcooL-graphql"
	"testing"
)

func TestAddProject(t *testing.T) {
	client := &mockGQLClient{
		mutateFunc: func(name string, mutation interface{}, variables map[string]interface{}) error {
			setNestedField(mutation, "AddProjectV2ItemById.Item.Id", "item-123")
			return nil
		},
	}

	itemId := addProject(client, "proj-1", "content-1")

	assertMutateCall(t, client, 0, "Assign")
	assertVariable(t, client.mutateCalls[0].Variables, "contentId", graphql.ID("content-1"))
	assertVariable(t, client.mutateCalls[0].Variables, "projectId", graphql.ID("proj-1"))

	if itemId != "item-123" {
		t.Errorf("itemId: got %q, want %q", itemId, "item-123")
	}
}

func TestUpdateTextProjectField(t *testing.T) {
	client := &mockGQLClient{}

	updateTextProjectField(client, "proj-1", "item-1", "field-1", "hello")

	assertMutateCall(t, client, 0, "UpdateFieldValue")
	vars := client.mutateCalls[0].Variables
	assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
	assertVariable(t, vars, "itemId", graphql.ID("item-1"))
	assertVariable(t, vars, "fieldId", graphql.ID("field-1"))
	assertVariable(t, vars, "value", graphql.String("hello"))
}

func TestUpdateNumberProjectField(t *testing.T) {
	client := &mockGQLClient{}

	updateNumberProjectField(client, "proj-1", "item-1", "field-1", 3.5)

	assertMutateCall(t, client, 0, "UpdateFieldValue")
	vars := client.mutateCalls[0].Variables
	assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
	assertVariable(t, vars, "itemId", graphql.ID("item-1"))
	assertVariable(t, vars, "fieldId", graphql.ID("field-1"))
	assertVariable(t, vars, "value", graphql.Float(3.5))
}

func TestUpdateSingleSelectProjectField(t *testing.T) {
	client := &mockGQLClient{}

	updateSingleSelectProjectField(client, "proj-1", "item-1", "field-1", "opt-abc")

	assertMutateCall(t, client, 0, "UpdateFieldValue")
	vars := client.mutateCalls[0].Variables
	assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
	assertVariable(t, vars, "itemId", graphql.ID("item-1"))
	assertVariable(t, vars, "fieldId", graphql.ID("field-1"))
	assertVariable(t, vars, "value", graphql.String("opt-abc"))
}

func TestUpdateIterationProjectField(t *testing.T) {
	client := &mockGQLClient{}

	updateIterationProjectField(client, "proj-1", "item-1", "field-1", "iter-abc")

	assertMutateCall(t, client, 0, "UpdateFieldValue")
	vars := client.mutateCalls[0].Variables
	assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
	assertVariable(t, vars, "itemId", graphql.ID("item-1"))
	assertVariable(t, vars, "fieldId", graphql.ID("field-1"))
	assertVariable(t, vars, "value", graphql.String("iter-abc"))
}

func TestUpdateDateProjectField(t *testing.T) {
	client := &mockGQLClient{}

	updateDateProjectField(client, "proj-1", "item-1", "field-1", "2024-03-15")

	assertMutateCall(t, client, 0, "UpdateFieldValue")
	vars := client.mutateCalls[0].Variables
	assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
	assertVariable(t, vars, "itemId", graphql.ID("item-1"))
	assertVariable(t, vars, "fieldId", graphql.ID("field-1"))

	// dateの値はgithubv4.Date型にパースされるので、文字列での比較はせず存在チェック
	if _, ok := vars["value"]; !ok {
		t.Error("variable 'value' not found")
	}
}
