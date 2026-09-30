package main

import "github.com/shurcooL/githubv4"

// GQLClient is an interface for GraphQL client operations, used for testability.
type GQLClient interface {
	Query(name string, q interface{}, variables map[string]interface{}) error
	Mutate(name string, m interface{}, variables map[string]interface{}) error
}

type Project struct {
	Title string
	Id    string
	Type  string
}

type Content struct {
	Id     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Option struct {
	Id   string
	Name string
}

type ProjectField struct {
	Id       string
	Name     string
	DataType string
	Options  []Option
}

type Repository struct {
	Name  string `json:"name"`
	Owner struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	} `json:"owner"`
	IsInOrganization bool `json:"isInOrganization"`
}
type NamedDateValue struct {
	Date githubv4.Date `json:"date,omitempty"`
}

type IterationOption struct {
	StartDate string
	Id        string
}

type SingleSelectOption struct {
	Id   string
	Name string
}

type ProjectFieldNode struct {
	ProjectV2IterationField struct {
		Id            string
		Name          string
		Configuration struct {
			Iterations []IterationOption
		}
	} `graphql:"... on ProjectV2IterationField"`
	ProjectV2SingleSelectField struct {
		Id      string
		Name    string
		Options []SingleSelectOption
	} `graphql:"... on ProjectV2SingleSelectField"`
}

type FieldType struct {
	Id       string
	Name     string
	DataType string
}

type ProjectItemFieldValueNode struct {
	Typename                    string `graphql:"__typename"`
	ProjectV2ItemFieldTextValue struct {
		Text  string
		Field struct {
			ProjectV2FieldCommon struct {
				Name string
			} `graphql:"... on ProjectV2FieldCommon"`
		}
	} `graphql:"... on ProjectV2ItemFieldTextValue"`
	ProjectV2ItemFieldNumberValue struct {
		Number float64
		Field  struct {
			ProjectV2FieldCommon struct {
				Name string
			} `graphql:"... on ProjectV2FieldCommon"`
		}
	} `graphql:"... on ProjectV2ItemFieldNumberValue"`
	ProjectV2ItemFieldDateValue struct {
		Date  string
		Field struct {
			ProjectV2FieldCommon struct {
				Name string
			} `graphql:"... on ProjectV2FieldCommon"`
		}
	} `graphql:"... on ProjectV2ItemFieldDateValue"`
	ProjectV2ItemFieldSingleSelectValue struct {
		Name  string
		Field struct {
			ProjectV2FieldCommon struct {
				Name string
			} `graphql:"... on ProjectV2FieldCommon"`
		}
	} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
	ProjectV2ItemFieldIterationValue struct {
		Title     string
		StartDate string
		Field     struct {
			ProjectV2FieldCommon struct {
				Name string
			} `graphql:"... on ProjectV2FieldCommon"`
		}
	} `graphql:"... on ProjectV2ItemFieldIterationValue"`
}

type ContentSummary struct {
	Number int
	Title  string
	Url    string
}

type ProjectItemNode struct {
	Project struct {
		Title  string
		Number int
		Url    string
	}
	Content struct {
		Typename    string         `graphql:"__typename"`
		Issue       ContentSummary `graphql:"... on Issue"`
		PullRequest ContentSummary `graphql:"... on PullRequest"`
	}
	FieldValues struct {
		Nodes []ProjectItemFieldValueNode
	} `graphql:"fieldValues(first: $number)"`
}

type FieldValue struct {
	Name  string
	Value string
}

type ItemSummary struct {
	ProjectTitle  string
	ProjectNumber int
	ProjectUrl    string
	ContentType   string
	ContentNumber int
	ContentTitle  string
	ContentUrl    string
	FieldValues   []FieldValue
}
