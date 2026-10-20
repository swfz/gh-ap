package main

import (
	"reflect"
	"testing"
)

func TestProjectOptionNames(t *testing.T) {
	tests := []struct {
		name        string
		projects    []Project
		wantNames   []string
		wantMissing int
	}{
		{
			name:        "空のリスト",
			projects:    []Project{},
			wantNames:   []string{},
			wantMissing: 0,
		},
		{
			name: "タイトルを順に並べる",
			projects: []Project{
				{Title: "Alpha", Id: "p1", Type: "UserProject"},
				{Title: "Beta", Id: "p2", Type: "OrganizationProject"},
			},
			wantNames:   []string{"Alpha", "Beta"},
			wantMissing: 0,
		},
		{
			name: "Id が空のプロジェクトを数える",
			projects: []Project{
				{Title: "Alpha", Id: ""},
				{Title: "Beta", Id: "p2"},
				{Title: "Gamma", Id: ""},
			},
			wantNames:   []string{"Alpha", "Beta", "Gamma"},
			wantMissing: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, missing := projectOptionNames(tt.projects)
			if !reflect.DeepEqual(names, tt.wantNames) {
				t.Errorf("names: got %v, want %v", names, tt.wantNames)
			}
			if missing != tt.wantMissing {
				t.Errorf("missing: got %d, want %d", missing, tt.wantMissing)
			}
		})
	}
}

func TestProjectDescription(t *testing.T) {
	projects := []Project{
		{Title: "Alpha", Id: "p1", Type: "UserProject"},
		{Title: "Beta", Id: "p2", Type: "OrganizationProject"},
	}
	description := projectDescription(projects)

	if got := description("Alpha", 0); got != "UserProject" {
		t.Errorf("index 0: got %q", got)
	}
	if got := description("Beta", 1); got != "OrganizationProject" {
		t.Errorf("index 1: got %q", got)
	}
}

func TestContentNumberOptions(t *testing.T) {
	tests := []struct {
		name     string
		contents []Content
		want     []string
	}{
		{name: "空のリスト", contents: []Content{}, want: []string{}},
		{
			name:     "番号を文字列にする",
			contents: []Content{{Number: 12, Title: "a"}, {Number: 3, Title: "b"}},
			want:     []string{"12", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contentNumberOptions(tt.contents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentDescription(t *testing.T) {
	contents := []Content{{Number: 1, Title: "Fix bug"}, {Number: 2, Title: "Add feature"}}
	description := contentDescription(contents)

	if got := description("1", 0); got != "Fix bug" {
		t.Errorf("index 0: got %q", got)
	}
	if got := description("2", 1); got != "Add feature" {
		t.Errorf("index 1: got %q", got)
	}
}

func TestContentFilter(t *testing.T) {
	contents := []Content{{Number: 1, Title: "Fix login bug"}, {Number: 2, Title: "Add feature"}}
	filter := contentFilter(contents)

	tests := []struct {
		name        string
		filterValue string
		optValue    string
		optIndex    int
		want        bool
	}{
		{name: "タイトルに部分一致", filterValue: "login", optValue: "1", optIndex: 0, want: true},
		{name: "タイトルに一致しない", filterValue: "login", optValue: "2", optIndex: 1, want: false},
		{name: "番号ではなくタイトルで判定する", filterValue: "2", optValue: "2", optIndex: 1, want: false},
		{name: "大文字小文字を区別する", filterValue: "fix", optValue: "1", optIndex: 0, want: false},
		{name: "空文字は全て一致", filterValue: "", optValue: "2", optIndex: 1, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter(tt.filterValue, tt.optValue, tt.optIndex); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectOptionNames(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    []string
	}{
		{name: "空のリスト", options: []Option{}, want: []string{}},
		{
			name:    "名前を順に並べる",
			options: []Option{{Id: "Skip", Name: "Skip This Question."}, {Id: "opt-1", Name: "Done"}},
			want:    []string{"Skip This Question.", "Done"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectOptionNames(tt.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "YYYY-MM-DD", value: "2026-10-21", wantErr: false},
		{name: "空文字は未入力として許可", value: "", wantErr: false},
		{name: "スラッシュ区切り", value: "2026/10/21", wantErr: true},
		{name: "存在しない日付", value: "2026-02-30", wantErr: true},
		{name: "日付でない文字列", value: "tomorrow", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err: got %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "Please format it like this '2006-01-02'." {
				t.Errorf("error message: got %q", err.Error())
			}
		})
	}
}

func TestValidateNumber(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "整数", value: "3", wantErr: false},
		{name: "小数", value: "1.5", wantErr: false},
		{name: "負数", value: "-2", wantErr: false},
		{name: "空文字", value: "", wantErr: true},
		{name: "数値でない文字列", value: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNumber(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err: got %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "Value is Int or Float." {
				t.Errorf("error message: got %q", err.Error())
			}
		})
	}
}
