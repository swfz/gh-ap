package main

import "github.com/cli/go-gh/pkg/api"

func getProjectFieldOptions(gqlclient api.GQLClient, projectId string) (fields []ProjectField) {
	nodes := queryProjectField(gqlclient, projectId)

	var fieldOptions []ProjectField
	for _, node := range nodes {
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

				field := ProjectField{
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

				field := ProjectField{
					Id:      node.ProjectV2SingleSelectField.Id,
					Name:    node.ProjectV2IterationField.Name,
					Options: iterationOptions,
				}
				fieldOptions = append(fieldOptions, field)
			}
		}
	}

	return fieldOptions
}

func getProjectFields(gqlclient api.GQLClient, projectId string) (fields []ProjectField) {
	fieldOptions := getProjectFieldOptions(gqlclient, projectId)
	fieldTypes := queryProjectFieldTypes(gqlclient, projectId)

	for _, fieldType := range fieldTypes {
		skipOption := []Option{{
			Id:   "Skip",
			Name: "Skip This Question.",
		},
		}
		field := ProjectField{
			Id:       fieldType.Id,
			Name:     fieldType.Name,
			DataType: fieldType.DataType,
		}
		if fieldType.DataType == "ITERATION" || fieldType.DataType == "SINGLE_SELECT" {
			for _, options := range fieldOptions {
				if options.Id == fieldType.Id {

					field.Options = append(skipOption, options.Options...)
					break
				}
			}
		} else {
			field.Options = []Option{}
		}
		fields = append(fields, field)
	}

	return fields
}
