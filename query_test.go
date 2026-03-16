package main

import (
	graphql "github.com/cli/shurcooL-graphql"
	"testing"
)

func TestQueryUserProjects(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			type node struct {
				Title string
				Id    string
			}
			nodes := []node{
				{Title: "Project A", Id: "proj-1"},
				{Title: "Project B", Id: "proj-2"},
			}
			setNestedField(query, "User.ProjectsV2.Nodes", nodes)
			return nil
		},
	}

	projects := queryUserProjects(client, "testuser")

	assertQueryCall(t, client, 0, "ProjectsV2")
	assertVariable(t, client.queryCalls[0].Variables, "login", graphql.String("testuser"))
	assertVariable(t, client.queryCalls[0].Variables, "size", graphql.Int(10))

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Title != "Project A" || projects[0].Id != "proj-1" {
		t.Errorf("unexpected project[0]: %+v", projects[0])
	}
	if projects[1].Title != "Project B" || projects[1].Id != "proj-2" {
		t.Errorf("unexpected project[1]: %+v", projects[1])
	}
}

func TestQueryOrganizationProjects(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			type node struct {
				Title string
				Id    string
			}
			nodes := []node{
				{Title: "Org Project", Id: "org-proj-1"},
			}
			setNestedField(query, "Organization.ProjectsV2.Nodes", nodes)
			return nil
		},
	}

	projects := queryOrganizationProjects(client, "myorg")

	assertQueryCall(t, client, 0, "ProjectsV2")
	assertVariable(t, client.queryCalls[0].Variables, "login", graphql.String("myorg"))
	assertVariable(t, client.queryCalls[0].Variables, "size", graphql.Int(10))

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Title != "Org Project" || projects[0].Id != "org-proj-1" {
		t.Errorf("unexpected project: %+v", projects[0])
	}
}

func TestQueryProjectFieldTypes(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			type fieldCommon struct {
				Id       string
				Name     string
				DataType string
			}
			type node struct {
				ProjectV2FieldCommon fieldCommon `graphql:"... on ProjectV2FieldCommon"`
			}
			nodes := []node{
				{ProjectV2FieldCommon: fieldCommon{Id: "f1", Name: "Title", DataType: "TEXT"}},
				{ProjectV2FieldCommon: fieldCommon{Id: "f2", Name: "Status", DataType: "SINGLE_SELECT"}},
				{ProjectV2FieldCommon: fieldCommon{Id: "f3", Name: "Points", DataType: "NUMBER"}},
			}
			setNestedField(query, "Node.ProjectV2.Fields.Nodes", nodes)
			return nil
		},
	}

	fieldTypes := queryProjectFieldTypes(client, "proj-123")

	assertQueryCall(t, client, 0, "FieldTypes")
	assertVariable(t, client.queryCalls[0].Variables, "projectId", graphql.ID("proj-123"))
	assertVariable(t, client.queryCalls[0].Variables, "number", graphql.Int(20))

	if len(fieldTypes) != 3 {
		t.Fatalf("expected 3 field types, got %d", len(fieldTypes))
	}

	tests := []struct {
		wantId       string
		wantName     string
		wantDataType string
	}{
		{"f1", "Title", "TEXT"},
		{"f2", "Status", "SINGLE_SELECT"},
		{"f3", "Points", "NUMBER"},
	}
	for i, tt := range tests {
		if fieldTypes[i].Id != tt.wantId {
			t.Errorf("[%d] Id: got %q, want %q", i, fieldTypes[i].Id, tt.wantId)
		}
		if fieldTypes[i].Name != tt.wantName {
			t.Errorf("[%d] Name: got %q, want %q", i, fieldTypes[i].Name, tt.wantName)
		}
		if fieldTypes[i].DataType != tt.wantDataType {
			t.Errorf("[%d] DataType: got %q, want %q", i, fieldTypes[i].DataType, tt.wantDataType)
		}
	}
}

func TestQueryProjectField(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			nodes := []ProjectFieldNode{
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id:   "ss-1",
						Name: "Status",
						Options: []SingleSelectOption{
							{Id: "opt-1", Name: "Todo"},
							{Id: "opt-2", Name: "Done"},
						},
					},
				},
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id: "iter-ss-1",
					},
					ProjectV2IterationField: struct {
						Id            string
						Name          string
						Configuration struct {
							Iterations []IterationOption
						}
					}{
						Name: "Sprint",
						Configuration: struct {
							Iterations []IterationOption
						}{
							Iterations: []IterationOption{
								{StartDate: "2024-01-01", Id: "iter-1"},
							},
						},
					},
				},
			}
			setNestedField(query, "Node.ProjectV2.Fields.Nodes", nodes)
			return nil
		},
	}

	result := queryProjectField(client, "proj-456")

	assertQueryCall(t, client, 0, "Fields")
	assertVariable(t, client.queryCalls[0].Variables, "projectId", graphql.ID("proj-456"))
	assertVariable(t, client.queryCalls[0].Variables, "number", graphql.Int(20))

	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}

	// SingleSelect node
	if result[0].ProjectV2SingleSelectField.Id != "ss-1" {
		t.Errorf("node[0] SingleSelect Id: got %q, want %q", result[0].ProjectV2SingleSelectField.Id, "ss-1")
	}
	if result[0].ProjectV2SingleSelectField.Name != "Status" {
		t.Errorf("node[0] SingleSelect Name: got %q, want %q", result[0].ProjectV2SingleSelectField.Name, "Status")
	}
	if len(result[0].ProjectV2SingleSelectField.Options) != 2 {
		t.Fatalf("node[0] SingleSelect Options: got %d, want 2", len(result[0].ProjectV2SingleSelectField.Options))
	}
	if result[0].ProjectV2SingleSelectField.Options[0].Id != "opt-1" || result[0].ProjectV2SingleSelectField.Options[0].Name != "Todo" {
		t.Errorf("node[0] option[0]: got %+v", result[0].ProjectV2SingleSelectField.Options[0])
	}

	// Iteration node
	if result[1].ProjectV2IterationField.Name != "Sprint" {
		t.Errorf("node[1] Iteration Name: got %q, want %q", result[1].ProjectV2IterationField.Name, "Sprint")
	}
	iters := result[1].ProjectV2IterationField.Configuration.Iterations
	if len(iters) != 1 {
		t.Fatalf("node[1] iterations: got %d, want 1", len(iters))
	}
	if iters[0].Id != "iter-1" || iters[0].StartDate != "2024-01-01" {
		t.Errorf("node[1] iteration[0]: got %+v", iters[0])
	}
}
