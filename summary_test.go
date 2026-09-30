package main

import (
	"bytes"
	"reflect"
	"testing"
)

func fieldValueNode(typename string, fieldName string) ProjectItemFieldValueNode {
	var node ProjectItemFieldValueNode
	node.Typename = typename
	// shurcooLのデコードでは同名キー(field)が全フラグメントに埋まるため、それを再現する
	node.ProjectV2ItemFieldTextValue.Field.ProjectV2FieldCommon.Name = fieldName
	node.ProjectV2ItemFieldNumberValue.Field.ProjectV2FieldCommon.Name = fieldName
	node.ProjectV2ItemFieldDateValue.Field.ProjectV2FieldCommon.Name = fieldName
	node.ProjectV2ItemFieldSingleSelectValue.Field.ProjectV2FieldCommon.Name = fieldName
	node.ProjectV2ItemFieldIterationValue.Field.ProjectV2FieldCommon.Name = fieldName
	return node
}

func TestBuildItemSummary(t *testing.T) {
	t.Run("PullRequestと各フィールド型", func(t *testing.T) {
		var node ProjectItemNode
		node.Project.Title = "individual-project"
		node.Project.Number = 2
		node.Project.Url = "https://github.com/users/swfz/projects/2"
		node.Content.Typename = "PullRequest"
		node.Content.PullRequest = ContentSummary{Number: 41, Title: "feat: something", Url: "https://github.com/swfz/repo/pull/41"}

		text := fieldValueNode("ProjectV2ItemFieldTextValue", "Title")
		text.ProjectV2ItemFieldTextValue.Text = "feat: something"
		number := fieldValueNode("ProjectV2ItemFieldNumberValue", "Point")
		number.ProjectV2ItemFieldNumberValue.Number = 1
		date := fieldValueNode("ProjectV2ItemFieldDateValue", "Month")
		date.ProjectV2ItemFieldDateValue.Date = "2026-10-01"
		selectValue := fieldValueNode("ProjectV2ItemFieldSingleSelectValue", "Status")
		selectValue.ProjectV2ItemFieldSingleSelectValue.Name = "Review"
		iteration := fieldValueNode("ProjectV2ItemFieldIterationValue", "Iteration")
		iteration.ProjectV2ItemFieldIterationValue.Title = "2026-10"
		iteration.ProjectV2ItemFieldIterationValue.StartDate = "2026-10-01"
		node.FieldValues.Nodes = []ProjectItemFieldValueNode{text, number, date, selectValue, iteration}

		got := buildItemSummary(node)
		want := ItemSummary{
			ProjectTitle:  "individual-project",
			ProjectNumber: 2,
			ProjectUrl:    "https://github.com/users/swfz/projects/2",
			ContentType:   "PullRequest",
			ContentNumber: 41,
			ContentTitle:  "feat: something",
			ContentUrl:    "https://github.com/swfz/repo/pull/41",
			FieldValues: []FieldValue{
				{Name: "Title", Value: "feat: something"},
				{Name: "Point", Value: "1"},
				{Name: "Month", Value: "2026-10-01"},
				{Name: "Status", Value: "Review"},
				{Name: "Iteration", Value: "2026-10 (2026-10-01)"},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("Issue", func(t *testing.T) {
		var node ProjectItemNode
		node.Content.Typename = "Issue"
		node.Content.Issue = ContentSummary{Number: 7, Title: "bug", Url: "https://github.com/swfz/repo/issues/7"}

		got := buildItemSummary(node)
		if got.ContentType != "Issue" || got.ContentNumber != 7 || got.ContentTitle != "bug" || got.ContentUrl != "https://github.com/swfz/repo/issues/7" {
			t.Errorf("unexpected content: %+v", got)
		}
		if len(got.FieldValues) != 0 {
			t.Errorf("expected no field values, got %+v", got.FieldValues)
		}
	})

	t.Run("小数を含む数値フィールド", func(t *testing.T) {
		var node ProjectItemNode
		number := fieldValueNode("ProjectV2ItemFieldNumberValue", "Point")
		number.ProjectV2ItemFieldNumberValue.Number = 2.5
		node.FieldValues.Nodes = []ProjectItemFieldValueNode{number}

		got := buildItemSummary(node)
		if len(got.FieldValues) != 1 || got.FieldValues[0].Value != "2.5" {
			t.Errorf("unexpected field values: %+v", got.FieldValues)
		}
	})

	// リグレッション: 全フラグメントにfield名が埋まっていても__typenameに対応する値を採用する
	t.Run("__typenameに応じた値を採用する", func(t *testing.T) {
		var node ProjectItemNode
		selectValue := fieldValueNode("ProjectV2ItemFieldSingleSelectValue", "Status")
		selectValue.ProjectV2ItemFieldSingleSelectValue.Name = "Done"
		node.FieldValues.Nodes = []ProjectItemFieldValueNode{selectValue}

		got := buildItemSummary(node)
		want := []FieldValue{{Name: "Status", Value: "Done"}}
		if !reflect.DeepEqual(got.FieldValues, want) {
			t.Errorf("got %+v, want %+v", got.FieldValues, want)
		}
	})

	t.Run("未対応の型は無視する", func(t *testing.T) {
		var node ProjectItemNode
		node.FieldValues.Nodes = []ProjectItemFieldValueNode{
			fieldValueNode("ProjectV2ItemFieldLabelValue", "Labels"),
			{},
		}

		got := buildItemSummary(node)
		if len(got.FieldValues) != 0 {
			t.Errorf("expected no field values, got %+v", got.FieldValues)
		}
	})
}

func TestPrintItemSummary(t *testing.T) {
	summary := ItemSummary{
		ProjectTitle:  "individual-project",
		ProjectNumber: 2,
		ProjectUrl:    "https://github.com/users/swfz/projects/2",
		ContentType:   "PullRequest",
		ContentNumber: 41,
		ContentTitle:  "feat: something",
		ContentUrl:    "https://github.com/swfz/repo/pull/41",
		FieldValues: []FieldValue{
			{Name: "Status", Value: "Review"},
			{Name: "Point", Value: "1"},
		},
	}

	var buf bytes.Buffer
	printItemSummary(&buf, summary)

	want := `Project: individual-project (#2) https://github.com/users/swfz/projects/2
Item: PullRequest #41 feat: something https://github.com/swfz/repo/pull/41
Fields:
  Status: Review
  Point: 1
`
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}
