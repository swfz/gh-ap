package main

import (
	"testing"
)

func TestBuildProjectFieldOptions(t *testing.T) {
	tests := []struct {
		name   string
		nodes  []ProjectFieldNode
		want   []ProjectField
	}{
		{
			name:  "空のノードでは空を返す",
			nodes: []ProjectFieldNode{},
			want:  nil,
		},
		{
			name: "SingleSelectフィールドのオプションを取得できる",
			nodes: []ProjectFieldNode{
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id:   "field-1",
						Name: "Status",
						Options: []SingleSelectOption{
							{Id: "opt-1", Name: "Todo"},
							{Id: "opt-2", Name: "Done"},
						},
					},
				},
			},
			want: []ProjectField{
				{
					Id:   "field-1",
					Name: "Status",
					Options: []Option{
						{Id: "opt-1", Name: "Todo"},
						{Id: "opt-2", Name: "Done"},
					},
				},
			},
		},
		{
			name: "オプションが空のSingleSelectはスキップされる",
			nodes: []ProjectFieldNode{
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id:   "field-1",
						Name: "EmptySelect",
					},
				},
			},
			want: nil,
		},
		{
			name: "Iterationフィールドのオプションを取得できる",
			nodes: []ProjectFieldNode{
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id: "iter-field-1",
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
								{StartDate: "2024-01-15", Id: "iter-2"},
							},
						},
					},
				},
			},
			want: []ProjectField{
				{
					Id:   "iter-field-1",
					Name: "Sprint",
					Options: []Option{
						{Id: "iter-1", Name: "2024-01-01"},
						{Id: "iter-2", Name: "2024-01-15"},
					},
				},
			},
		},
		{
			name: "SingleSelectとIterationが混在する場合は両方返す",
			nodes: []ProjectFieldNode{
				{
					ProjectV2SingleSelectField: struct {
						Id      string
						Name    string
						Options []SingleSelectOption
					}{
						Id:   "ss-1",
						Name: "Status",
						Options: []SingleSelectOption{
							{Id: "opt-1", Name: "Open"},
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
								{StartDate: "2024-03-01", Id: "iter-id-1"},
							},
						},
					},
				},
			},
			want: []ProjectField{
				{
					Id:   "ss-1",
					Name: "Status",
					Options: []Option{
						{Id: "opt-1", Name: "Open"},
					},
				},
				{
					Id:   "iter-ss-1",
					Name: "Sprint",
					Options: []Option{
						{Id: "iter-id-1", Name: "2024-03-01"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildProjectFieldOptions(tt.nodes)
			assertProjectFields(t, got, tt.want)
		})
	}
}

func TestMergeFieldsWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		fieldOptions []ProjectField
		fieldTypes   []FieldType
		want         []ProjectField
	}{
		{
			name:         "TEXTフィールドは空のOptionsを持つ",
			fieldOptions: nil,
			fieldTypes: []FieldType{
				{Id: "text-1", Name: "Description", DataType: "TEXT"},
			},
			want: []ProjectField{
				{Id: "text-1", Name: "Description", DataType: "TEXT", Options: []Option{}},
			},
		},
		{
			name:         "DATEフィールドは空のOptionsを持つ",
			fieldOptions: nil,
			fieldTypes: []FieldType{
				{Id: "date-1", Name: "DueDate", DataType: "DATE"},
			},
			want: []ProjectField{
				{Id: "date-1", Name: "DueDate", DataType: "DATE", Options: []Option{}},
			},
		},
		{
			name:         "NUMBERフィールドは空のOptionsを持つ",
			fieldOptions: nil,
			fieldTypes: []FieldType{
				{Id: "num-1", Name: "Points", DataType: "NUMBER"},
			},
			want: []ProjectField{
				{Id: "num-1", Name: "Points", DataType: "NUMBER", Options: []Option{}},
			},
		},
		{
			name: "SINGLE_SELECTフィールドはSkipオプションが先頭に追加される",
			fieldOptions: []ProjectField{
				{
					Id: "ss-field",
					Options: []Option{
						{Id: "p1", Name: "High"},
						{Id: "p2", Name: "Medium"},
						{Id: "p3", Name: "Low"},
					},
				},
			},
			fieldTypes: []FieldType{
				{Id: "ss-field", Name: "Priority", DataType: "SINGLE_SELECT"},
			},
			want: []ProjectField{
				{
					Id: "ss-field", Name: "Priority", DataType: "SINGLE_SELECT",
					Options: []Option{
						{Id: "Skip", Name: "Skip This Question."},
						{Id: "p1", Name: "High"},
						{Id: "p2", Name: "Medium"},
						{Id: "p3", Name: "Low"},
					},
				},
			},
		},
		{
			name: "ITERATIONフィールドもSkipオプションが先頭に追加される",
			fieldOptions: []ProjectField{
				{
					Id: "iter-field",
					Options: []Option{
						{Id: "iter-1", Name: "2024-01-01"},
					},
				},
			},
			fieldTypes: []FieldType{
				{Id: "iter-field", Name: "Sprint", DataType: "ITERATION"},
			},
			want: []ProjectField{
				{
					Id: "iter-field", Name: "Sprint", DataType: "ITERATION",
					Options: []Option{
						{Id: "Skip", Name: "Skip This Question."},
						{Id: "iter-1", Name: "2024-01-01"},
					},
				},
			},
		},
		{
			name:         "SINGLE_SELECTでfieldOptionsにマッチがない場合Optionsはnil",
			fieldOptions: nil,
			fieldTypes: []FieldType{
				{Id: "ss-field", Name: "Status", DataType: "SINGLE_SELECT"},
			},
			want: []ProjectField{
				{Id: "ss-field", Name: "Status", DataType: "SINGLE_SELECT", Options: nil},
			},
		},
		{
			name: "全フィールドタイプが混在する場合",
			fieldOptions: []ProjectField{
				{
					Id:      "status-field",
					Options: []Option{{Id: "opt-todo", Name: "Todo"}},
				},
			},
			fieldTypes: []FieldType{
				{Id: "text-field", Name: "Description", DataType: "TEXT"},
				{Id: "date-field", Name: "DueDate", DataType: "DATE"},
				{Id: "num-field", Name: "Points", DataType: "NUMBER"},
				{Id: "status-field", Name: "Status", DataType: "SINGLE_SELECT"},
				{Id: "iter-field", Name: "Sprint", DataType: "ITERATION"},
			},
			want: []ProjectField{
				{Id: "text-field", Name: "Description", DataType: "TEXT", Options: []Option{}},
				{Id: "date-field", Name: "DueDate", DataType: "DATE", Options: []Option{}},
				{Id: "num-field", Name: "Points", DataType: "NUMBER", Options: []Option{}},
				{
					Id: "status-field", Name: "Status", DataType: "SINGLE_SELECT",
					Options: []Option{
						{Id: "Skip", Name: "Skip This Question."},
						{Id: "opt-todo", Name: "Todo"},
					},
				},
				{Id: "iter-field", Name: "Sprint", DataType: "ITERATION", Options: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeFieldsWithOptions(tt.fieldOptions, tt.fieldTypes)
			assertProjectFields(t, got, tt.want)
		})
	}
}

func assertProjectFields(t *testing.T, got, want []ProjectField) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i].Id != want[i].Id {
			t.Errorf("[%d] Id: got %q, want %q", i, got[i].Id, want[i].Id)
		}
		if got[i].Name != want[i].Name {
			t.Errorf("[%d] Name: got %q, want %q", i, got[i].Name, want[i].Name)
		}
		if got[i].DataType != want[i].DataType {
			t.Errorf("[%d] DataType: got %q, want %q", i, got[i].DataType, want[i].DataType)
		}
		assertOptions(t, i, got[i].Options, want[i].Options)
	}
}

func assertOptions(t *testing.T, fieldIdx int, got, want []Option) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("[%d] Options: got %v, want nil", fieldIdx, got)
		}
		return
	}

	if len(got) != len(want) {
		t.Fatalf("[%d] Options len: got %d, want %d", fieldIdx, len(got), len(want))
	}

	for j := range want {
		if got[j].Id != want[j].Id || got[j].Name != want[j].Name {
			t.Errorf("[%d] Options[%d]: got %+v, want %+v", fieldIdx, j, got[j], want[j])
		}
	}
}
