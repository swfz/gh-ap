package main

import (
	"context"
	"reflect"
	"testing"
)

type mockGQLClient struct {
	queryFunc  func(name string, query interface{}, variables map[string]interface{}) error
	mutateFunc func(name string, mutation interface{}, variables map[string]interface{}) error
}

func (m *mockGQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	return nil
}

func (m *mockGQLClient) DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	return nil
}

func (m *mockGQLClient) Mutate(name string, mutation interface{}, variables map[string]interface{}) error {
	if m.mutateFunc != nil {
		return m.mutateFunc(name, mutation, variables)
	}
	return nil
}

func (m *mockGQLClient) MutateWithContext(ctx context.Context, name string, mutation interface{}, variables map[string]interface{}) error {
	return nil
}

func (m *mockGQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	if m.queryFunc != nil {
		return m.queryFunc(name, query, variables)
	}
	return nil
}

func (m *mockGQLClient) QueryWithContext(ctx context.Context, name string, query interface{}, variables map[string]interface{}) error {
	return nil
}

func TestGetProjectFieldOptions_SingleSelect(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNode := reflect.New(nodeType).Elem()

				ssField := newNode.FieldByName("ProjectV2SingleSelectField")
				ssField.FieldByName("Id").SetString("field-1")
				ssField.FieldByName("Name").SetString("Status")

				opts := ssField.FieldByName("Options")
				optType := opts.Type().Elem()

				opt1 := reflect.New(optType).Elem()
				opt1.FieldByName("Id").SetString("opt-1")
				opt1.FieldByName("Name").SetString("Todo")

				opt2 := reflect.New(optType).Elem()
				opt2.FieldByName("Id").SetString("opt-2")
				opt2.FieldByName("Name").SetString("Done")

				opts = reflect.Append(opts, opt1, opt2)
				ssField.FieldByName("Options").Set(opts)

				nodes = reflect.Append(nodes, newNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFieldOptions(client, "project-1")

	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].Id != "field-1" {
		t.Errorf("expected field id 'field-1', got '%s'", fields[0].Id)
	}
	if fields[0].Name != "Status" {
		t.Errorf("expected field name 'Status', got '%s'", fields[0].Name)
	}
	if len(fields[0].Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(fields[0].Options))
	}
	if fields[0].Options[0].Id != "opt-1" || fields[0].Options[0].Name != "Todo" {
		t.Errorf("unexpected first option: %+v", fields[0].Options[0])
	}
	if fields[0].Options[1].Id != "opt-2" || fields[0].Options[1].Name != "Done" {
		t.Errorf("unexpected second option: %+v", fields[0].Options[1])
	}
}

func TestGetProjectFieldOptions_Iteration(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNode := reflect.New(nodeType).Elem()

				// Set SingleSelectField Id (the current code checks this for iteration too)
				ssField := newNode.FieldByName("ProjectV2SingleSelectField")
				ssField.FieldByName("Id").SetString("iter-field-1")

				// Set IterationField
				iterField := newNode.FieldByName("ProjectV2IterationField")
				iterField.FieldByName("Name").SetString("Sprint")

				config := iterField.FieldByName("Configuration")
				iters := config.FieldByName("Iterations")
				iterType := iters.Type().Elem()

				iter1 := reflect.New(iterType).Elem()
				iter1.FieldByName("StartDate").SetString("2024-01-01")
				iter1.FieldByName("Id").SetString("iter-1")

				iter2 := reflect.New(iterType).Elem()
				iter2.FieldByName("StartDate").SetString("2024-01-15")
				iter2.FieldByName("Id").SetString("iter-2")

				iters = reflect.Append(iters, iter1, iter2)
				config.FieldByName("Iterations").Set(iters)

				nodes = reflect.Append(nodes, newNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFieldOptions(client, "project-1")

	// Current behavior: SingleSelect with no options is skipped, but iteration with iterations is added
	// The code checks ProjectV2SingleSelectField.Id != "" for both blocks
	if len(fields) != 1 {
		t.Fatalf("expected 1 field (iteration), got %d", len(fields))
	}
	if fields[0].Name != "Sprint" {
		t.Errorf("expected field name 'Sprint', got '%s'", fields[0].Name)
	}
	// Iteration field uses SingleSelectField.Id as its Id (current behavior)
	if fields[0].Id != "iter-field-1" {
		t.Errorf("expected field id 'iter-field-1', got '%s'", fields[0].Id)
	}
	if len(fields[0].Options) != 2 {
		t.Fatalf("expected 2 iteration options, got %d", len(fields[0].Options))
	}
	if fields[0].Options[0].Id != "iter-1" || fields[0].Options[0].Name != "2024-01-01" {
		t.Errorf("unexpected first iteration option: %+v", fields[0].Options[0])
	}
}

func TestGetProjectFieldOptions_EmptyNodes(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			return nil
		},
	}

	fields := getProjectFieldOptions(client, "project-1")

	if len(fields) != 0 {
		t.Fatalf("expected 0 fields, got %d", len(fields))
	}
}

func TestGetProjectFieldOptions_SingleSelectWithNoOptions(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNode := reflect.New(nodeType).Elem()

				ssField := newNode.FieldByName("ProjectV2SingleSelectField")
				ssField.FieldByName("Id").SetString("field-1")
				ssField.FieldByName("Name").SetString("EmptySelect")
				// Options left empty

				nodes = reflect.Append(nodes, newNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFieldOptions(client, "project-1")

	// SingleSelect with no options should be skipped
	if len(fields) != 0 {
		t.Fatalf("expected 0 fields for empty options, got %d", len(fields))
	}
}

func TestGetProjectFields_MixedTypes(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNode := reflect.New(nodeType).Elem()

				ssField := newNode.FieldByName("ProjectV2SingleSelectField")
				ssField.FieldByName("Id").SetString("status-field")
				ssField.FieldByName("Name").SetString("Status")

				opts := ssField.FieldByName("Options")
				optType := opts.Type().Elem()

				opt1 := reflect.New(optType).Elem()
				opt1.FieldByName("Id").SetString("opt-todo")
				opt1.FieldByName("Name").SetString("Todo")

				opts = reflect.Append(opts, opt1)
				ssField.FieldByName("Options").Set(opts)

				nodes = reflect.Append(nodes, newNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			if name == "FieldTypes" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()

				// TEXT field
				textNode := reflect.New(nodeType).Elem()
				common := textNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("text-field")
				common.FieldByName("Name").SetString("Description")
				common.FieldByName("DataType").SetString("TEXT")
				nodes = reflect.Append(nodes, textNode)

				// DATE field
				dateNode := reflect.New(nodeType).Elem()
				common = dateNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("date-field")
				common.FieldByName("Name").SetString("DueDate")
				common.FieldByName("DataType").SetString("DATE")
				nodes = reflect.Append(nodes, dateNode)

				// NUMBER field
				numNode := reflect.New(nodeType).Elem()
				common = numNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("num-field")
				common.FieldByName("Name").SetString("Points")
				common.FieldByName("DataType").SetString("NUMBER")
				nodes = reflect.Append(nodes, numNode)

				// SINGLE_SELECT field
				ssNode := reflect.New(nodeType).Elem()
				common = ssNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("status-field")
				common.FieldByName("Name").SetString("Status")
				common.FieldByName("DataType").SetString("SINGLE_SELECT")
				nodes = reflect.Append(nodes, ssNode)

				// ITERATION field
				iterNode := reflect.New(nodeType).Elem()
				common = iterNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("iter-field")
				common.FieldByName("Name").SetString("Sprint")
				common.FieldByName("DataType").SetString("ITERATION")
				nodes = reflect.Append(nodes, iterNode)

				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFields(client, "project-1")

	if len(fields) != 5 {
		t.Fatalf("expected 5 fields, got %d", len(fields))
	}

	// TEXT field
	if fields[0].Id != "text-field" || fields[0].Name != "Description" || fields[0].DataType != "TEXT" {
		t.Errorf("unexpected text field: %+v", fields[0])
	}
	if len(fields[0].Options) != 0 {
		t.Errorf("text field should have empty options, got %d", len(fields[0].Options))
	}

	// DATE field
	if fields[1].Id != "date-field" || fields[1].Name != "DueDate" || fields[1].DataType != "DATE" {
		t.Errorf("unexpected date field: %+v", fields[1])
	}
	if len(fields[1].Options) != 0 {
		t.Errorf("date field should have empty options, got %d", len(fields[1].Options))
	}

	// NUMBER field
	if fields[2].Id != "num-field" || fields[2].Name != "Points" || fields[2].DataType != "NUMBER" {
		t.Errorf("unexpected number field: %+v", fields[2])
	}
	if len(fields[2].Options) != 0 {
		t.Errorf("number field should have empty options, got %d", len(fields[2].Options))
	}

	// SINGLE_SELECT field - should have Skip option + actual options
	if fields[3].Id != "status-field" || fields[3].Name != "Status" || fields[3].DataType != "SINGLE_SELECT" {
		t.Errorf("unexpected single_select field: %+v", fields[3])
	}
	if len(fields[3].Options) != 2 {
		t.Fatalf("single_select field should have 2 options (Skip + Todo), got %d", len(fields[3].Options))
	}
	if fields[3].Options[0].Id != "Skip" || fields[3].Options[0].Name != "Skip This Question." {
		t.Errorf("first option should be Skip, got: %+v", fields[3].Options[0])
	}
	if fields[3].Options[1].Id != "opt-todo" || fields[3].Options[1].Name != "Todo" {
		t.Errorf("second option should be Todo, got: %+v", fields[3].Options[1])
	}

	// ITERATION field - no matching options from getProjectFieldOptions (different ID)
	if fields[4].Id != "iter-field" || fields[4].Name != "Sprint" || fields[4].DataType != "ITERATION" {
		t.Errorf("unexpected iteration field: %+v", fields[4])
	}
}

func TestGetProjectFields_SkipOptionPrepended(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				newNode := reflect.New(nodeType).Elem()

				ssField := newNode.FieldByName("ProjectV2SingleSelectField")
				ssField.FieldByName("Id").SetString("ss-field")
				ssField.FieldByName("Name").SetString("Priority")

				opts := ssField.FieldByName("Options")
				optType := opts.Type().Elem()

				opt1 := reflect.New(optType).Elem()
				opt1.FieldByName("Id").SetString("p1")
				opt1.FieldByName("Name").SetString("High")

				opt2 := reflect.New(optType).Elem()
				opt2.FieldByName("Id").SetString("p2")
				opt2.FieldByName("Name").SetString("Medium")

				opt3 := reflect.New(optType).Elem()
				opt3.FieldByName("Id").SetString("p3")
				opt3.FieldByName("Name").SetString("Low")

				opts = reflect.Append(opts, opt1, opt2, opt3)
				ssField.FieldByName("Options").Set(opts)

				nodes = reflect.Append(nodes, newNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			if name == "FieldTypes" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				ssNode := reflect.New(nodeType).Elem()
				common := ssNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("ss-field")
				common.FieldByName("Name").SetString("Priority")
				common.FieldByName("DataType").SetString("SINGLE_SELECT")

				nodes = reflect.Append(nodes, ssNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFields(client, "project-1")

	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}

	// Skip option should be first, followed by the actual options
	if len(fields[0].Options) != 4 {
		t.Fatalf("expected 4 options (Skip + 3), got %d", len(fields[0].Options))
	}
	if fields[0].Options[0].Id != "Skip" {
		t.Errorf("first option should be Skip, got: %s", fields[0].Options[0].Id)
	}
	if fields[0].Options[1].Id != "p1" {
		t.Errorf("second option should be p1, got: %s", fields[0].Options[1].Id)
	}
	if fields[0].Options[2].Id != "p2" {
		t.Errorf("third option should be p2, got: %s", fields[0].Options[2].Id)
	}
	if fields[0].Options[3].Id != "p3" {
		t.Errorf("fourth option should be p3, got: %s", fields[0].Options[3].Id)
	}
}

func TestGetProjectFields_NoSelectOptionsMatch(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				// Return no field options
			}
			if name == "FieldTypes" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()
				ssNode := reflect.New(nodeType).Elem()
				common := ssNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("ss-field")
				common.FieldByName("Name").SetString("Status")
				common.FieldByName("DataType").SetString("SINGLE_SELECT")

				nodes = reflect.Append(nodes, ssNode)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFields(client, "project-1")

	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}

	// SINGLE_SELECT with no matching fieldOptions: Options should be nil (no Skip prepended because no match found)
	if fields[0].Options != nil {
		t.Errorf("expected nil options when no field options match, got %+v", fields[0].Options)
	}
}

func TestGetProjectFields_OnlyNonSelectFields(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				// No select/iteration fields
			}
			if name == "FieldTypes" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()

				textNode := reflect.New(nodeType).Elem()
				common := textNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("text-1")
				common.FieldByName("Name").SetString("Title")
				common.FieldByName("DataType").SetString("TEXT")
				nodes = reflect.Append(nodes, textNode)

				dateNode := reflect.New(nodeType).Elem()
				common = dateNode.FieldByName("ProjectV2FieldCommon")
				common.FieldByName("Id").SetString("date-1")
				common.FieldByName("Name").SetString("Due")
				common.FieldByName("DataType").SetString("DATE")
				nodes = reflect.Append(nodes, dateNode)

				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFields(client, "project-1")

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	for _, f := range fields {
		if len(f.Options) != 0 {
			t.Errorf("non-select field %s should have empty options, got %d", f.Name, len(f.Options))
		}
	}
}

func TestGetProjectFieldOptions_BothSingleSelectAndIteration(t *testing.T) {
	client := &mockGQLClient{
		queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
			if name == "Fields" {
				v := reflect.ValueOf(query).Elem()
				nodes := v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes")

				nodeType := nodes.Type().Elem()

				// Node 1: SingleSelect with options
				node1 := reflect.New(nodeType).Elem()
				ssField1 := node1.FieldByName("ProjectV2SingleSelectField")
				ssField1.FieldByName("Id").SetString("ss-1")
				ssField1.FieldByName("Name").SetString("Status")
				opts := ssField1.FieldByName("Options")
				optType := opts.Type().Elem()
				opt := reflect.New(optType).Elem()
				opt.FieldByName("Id").SetString("opt-1")
				opt.FieldByName("Name").SetString("Open")
				opts = reflect.Append(opts, opt)
				ssField1.FieldByName("Options").Set(opts)

				// Node 2: Iteration field
				node2 := reflect.New(nodeType).Elem()
				ssField2 := node2.FieldByName("ProjectV2SingleSelectField")
				ssField2.FieldByName("Id").SetString("iter-ss-1")
				iterField := node2.FieldByName("ProjectV2IterationField")
				iterField.FieldByName("Name").SetString("Sprint")
				config := iterField.FieldByName("Configuration")
				iters := config.FieldByName("Iterations")
				iterType := iters.Type().Elem()
				iter := reflect.New(iterType).Elem()
				iter.FieldByName("StartDate").SetString("2024-03-01")
				iter.FieldByName("Id").SetString("iter-id-1")
				iters = reflect.Append(iters, iter)
				config.FieldByName("Iterations").Set(iters)

				nodes = reflect.Append(nodes, node1, node2)
				v.FieldByName("Node").FieldByName("ProjectV2").FieldByName("Fields").FieldByName("Nodes").Set(nodes)
			}
			return nil
		},
	}

	fields := getProjectFieldOptions(client, "project-1")

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	// First: SingleSelect
	if fields[0].Name != "Status" {
		t.Errorf("expected first field name 'Status', got '%s'", fields[0].Name)
	}
	if len(fields[0].Options) != 1 {
		t.Errorf("expected 1 option for Status, got %d", len(fields[0].Options))
	}

	// Second: Iteration
	if fields[1].Name != "Sprint" {
		t.Errorf("expected second field name 'Sprint', got '%s'", fields[1].Name)
	}
	if fields[1].Id != "iter-ss-1" {
		t.Errorf("expected iteration field id 'iter-ss-1', got '%s'", fields[1].Id)
	}
	if len(fields[1].Options) != 1 {
		t.Errorf("expected 1 option for Sprint, got %d", len(fields[1].Options))
	}
	if fields[1].Options[0].Name != "2024-03-01" {
		t.Errorf("expected iteration option name '2024-03-01', got '%s'", fields[1].Options[0].Name)
	}
}
