package main

import (
	"fmt"
	"io"
	"strconv"
)

func buildItemSummary(node ProjectItemNode) ItemSummary {
	summary := ItemSummary{
		ProjectTitle:  node.Project.Title,
		ProjectNumber: node.Project.Number,
		ProjectUrl:    node.Project.Url,
	}

	switch node.Content.Typename {
	case "PullRequest":
		summary.ContentType = "PullRequest"
		summary.ContentNumber = node.Content.PullRequest.Number
		summary.ContentTitle = node.Content.PullRequest.Title
		summary.ContentUrl = node.Content.PullRequest.Url
	case "Issue":
		summary.ContentType = "Issue"
		summary.ContentNumber = node.Content.Issue.Number
		summary.ContentTitle = node.Content.Issue.Title
		summary.ContentUrl = node.Content.Issue.Url
	}

	for _, v := range node.FieldValues.Nodes {
		switch v.Typename {
		case "ProjectV2ItemFieldTextValue":
			summary.FieldValues = append(summary.FieldValues, FieldValue{
				Name:  v.ProjectV2ItemFieldTextValue.Field.ProjectV2FieldCommon.Name,
				Value: v.ProjectV2ItemFieldTextValue.Text,
			})
		case "ProjectV2ItemFieldNumberValue":
			summary.FieldValues = append(summary.FieldValues, FieldValue{
				Name:  v.ProjectV2ItemFieldNumberValue.Field.ProjectV2FieldCommon.Name,
				Value: strconv.FormatFloat(v.ProjectV2ItemFieldNumberValue.Number, 'f', -1, 64),
			})
		case "ProjectV2ItemFieldDateValue":
			summary.FieldValues = append(summary.FieldValues, FieldValue{
				Name:  v.ProjectV2ItemFieldDateValue.Field.ProjectV2FieldCommon.Name,
				Value: v.ProjectV2ItemFieldDateValue.Date,
			})
		case "ProjectV2ItemFieldSingleSelectValue":
			summary.FieldValues = append(summary.FieldValues, FieldValue{
				Name:  v.ProjectV2ItemFieldSingleSelectValue.Field.ProjectV2FieldCommon.Name,
				Value: v.ProjectV2ItemFieldSingleSelectValue.Name,
			})
		case "ProjectV2ItemFieldIterationValue":
			summary.FieldValues = append(summary.FieldValues, FieldValue{
				Name:  v.ProjectV2ItemFieldIterationValue.Field.ProjectV2FieldCommon.Name,
				Value: v.ProjectV2ItemFieldIterationValue.Title + " (" + v.ProjectV2ItemFieldIterationValue.StartDate + ")",
			})
		}
	}

	return summary
}

func printItemSummary(w io.Writer, summary ItemSummary) {
	fmt.Fprintf(w, "Project: %s (#%d) %s\n", summary.ProjectTitle, summary.ProjectNumber, summary.ProjectUrl)
	fmt.Fprintf(w, "Item: %s #%d %s %s\n", summary.ContentType, summary.ContentNumber, summary.ContentTitle, summary.ContentUrl)
	fmt.Fprintln(w, "Fields:")
	for _, f := range summary.FieldValues {
		fmt.Fprintf(w, "  %s: %s\n", f.Name, f.Value)
	}
}
