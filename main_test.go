package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"testing"

	graphql "github.com/cli/shurcooL-graphql"
)

func TestParseFieldFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    fieldFlags
		expected map[string]string
	}{
		{
			name:     "空のフラグ",
			flags:    fieldFlags{},
			expected: map[string]string{},
		},
		{
			name:  "単一フィールド",
			flags: fieldFlags{"Status=Done"},
			expected: map[string]string{
				"Status": "Done",
			},
		},
		{
			name:  "複数フィールド",
			flags: fieldFlags{"Status=Done", "Priority=High"},
			expected: map[string]string{
				"Status":   "Done",
				"Priority": "High",
			},
		},
		{
			name:  "値にイコールを含む",
			flags: fieldFlags{"Note=a=b=c"},
			expected: map[string]string{
				"Note": "a=b=c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFieldFlags(tt.flags)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestFieldFlagsSet(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "正しいフォーマット",
			value:   "Status=Done",
			wantErr: false,
		},
		{
			name:    "イコールなし",
			value:   "InvalidFormat",
			wantErr: true,
		},
		{
			name:    "空の値",
			value:   "Key=",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f fieldFlags
			err := f.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%s) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if err == nil && len(f) != 1 {
				t.Errorf("expected 1 flag, got %d", len(f))
			}
		})
	}
}

func TestFindProjectByName(t *testing.T) {
	projects := []Project{
		{Title: "Project Alpha", Id: "id-alpha", Type: "UserProject"},
		{Title: "Project Beta", Id: "id-beta", Type: "OrganizationProject"},
		{Title: "My Board", Id: "id-board", Type: "UserProject"},
	}

	tests := []struct {
		name      string
		search    string
		wantFound bool
		wantId    string
	}{
		{
			name:      "存在するプロジェクト",
			search:    "Project Alpha",
			wantFound: true,
			wantId:    "id-alpha",
		},
		{
			name:      "存在しないプロジェクト",
			search:    "NotExist",
			wantFound: false,
		},
		{
			name:      "完全一致が必要",
			search:    "Project",
			wantFound: false,
		},
		{
			name:      "別のプロジェクト",
			search:    "My Board",
			wantFound: true,
			wantId:    "id-board",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found := findProjectByName(projects, tt.search)
			if found != tt.wantFound {
				t.Errorf("findProjectByName(%s) found = %v, want %v", tt.search, found, tt.wantFound)
			}
			if found && id != tt.wantId {
				t.Errorf("findProjectByName(%s) Id = %s, want %s", tt.search, id, tt.wantId)
			}
		})
	}
}

func TestFindOptionByName(t *testing.T) {
	options := []Option{
		{Id: "id1", Name: "Done"},
		{Id: "id2", Name: "In Progress"},
		{Id: "id3", Name: "Todo"},
	}

	tests := []struct {
		name      string
		search    string
		wantFound bool
		wantId    string
	}{
		{
			name:      "存在するオプション",
			search:    "Done",
			wantFound: true,
			wantId:    "id1",
		},
		{
			name:      "存在しないオプション",
			search:    "NotExist",
			wantFound: false,
		},
		{
			name:      "スペースを含むオプション",
			search:    "In Progress",
			wantFound: true,
			wantId:    "id2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, found := findOptionByName(options, tt.search)
			if found != tt.wantFound {
				t.Errorf("findOptionByName(%s) found = %v, want %v", tt.search, found, tt.wantFound)
			}
			if found && opt.Id != tt.wantId {
				t.Errorf("findOptionByName(%s) Id = %s, want %s", tt.search, opt.Id, tt.wantId)
			}
		})
	}
}

type mockRESTClient struct {
	login string
	paths []string
}

func (m *mockRESTClient) Get(path string, response interface{}) error {
	m.paths = append(m.paths, path)
	b, _ := json.Marshal(map[string]string{"login": m.login})
	return json.Unmarshal(b, response)
}

func repositoryFunc(t *testing.T, repository Repository, called *bool) func() Repository {
	t.Helper()
	return func() Repository {
		*called = true
		return repository
	}
}

func orgRepository(login string) Repository {
	r := Repository{IsInOrganization: true}
	r.Owner.Login = login
	return r
}

func TestGetProjects(t *testing.T) {
	newClient := func() *mockGQLClient {
		return &mockGQLClient{
			queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
				type node struct {
					Title string
					Id    string
				}
				if variables["login"] == graphql.String("testuser") {
					setNestedField(query, "User.ProjectsV2.Nodes", []node{{Title: "User A", Id: "user-1"}, {Title: "User B", Id: "user-2"}})
				} else {
					setNestedField(query, "Organization.ProjectsV2.Nodes", []node{{Title: "Org A", Id: "org-1"}})
				}
				return nil
			},
		}
	}

	t.Run("組織のリポジトリではユーザーと組織のプロジェクトを重複なく結合する", func(t *testing.T) {
		client := newClient()
		rest := &mockRESTClient{login: "testuser"}
		var called bool

		projects := getProjects(client, rest, repositoryFunc(t, orgRepository("myorg"), &called))

		want := []Project{
			{Title: "User A", Id: "user-1", Type: "UserProject"},
			{Title: "User B", Id: "user-2", Type: "UserProject"},
			{Title: "Org A", Id: "org-1", Type: "OrganizationProject"},
		}
		if !reflect.DeepEqual(projects, want) {
			t.Errorf("got %+v, want %+v", projects, want)
		}
		if !reflect.DeepEqual(rest.paths, []string{"user"}) {
			t.Errorf("rest paths: got %v", rest.paths)
		}
		if len(client.queryCalls) != 2 {
			t.Fatalf("expected 2 query calls, got %d", len(client.queryCalls))
		}
		assertVariable(t, client.queryCalls[0].Variables, "login", graphql.String("testuser"))
		assertVariable(t, client.queryCalls[1].Variables, "login", graphql.String("myorg"))
	})

	t.Run("個人のリポジトリではユーザーのプロジェクトのみ", func(t *testing.T) {
		client := newClient()
		var called bool

		projects := getProjects(client, &mockRESTClient{login: "testuser"}, repositoryFunc(t, Repository{}, &called))

		want := []Project{
			{Title: "User A", Id: "user-1", Type: "UserProject"},
			{Title: "User B", Id: "user-2", Type: "UserProject"},
		}
		if !reflect.DeepEqual(projects, want) {
			t.Errorf("got %+v, want %+v", projects, want)
		}
		if len(client.queryCalls) != 1 {
			t.Errorf("expected 1 query call, got %d", len(client.queryCalls))
		}
	})
}

func TestResolveProjectIdByNumber(t *testing.T) {
	newClient := func(userId, orgId string) *mockGQLClient {
		return &mockGQLClient{
			queryFunc: func(name string, query interface{}, variables map[string]interface{}) error {
				if name == "UserProjectV2" && userId != "" {
					setNestedField(query, "User.ProjectV2.Id", userId)
				}
				if name == "OrgProjectV2" && orgId != "" {
					setNestedField(query, "Organization.ProjectV2.Id", orgId)
				}
				return nil
			},
		}
	}

	t.Run("ユーザーのプロジェクトで見つかれば組織は探さない", func(t *testing.T) {
		client := newClient("PVT_user", "PVT_org")
		var called bool

		id, found := resolveProjectIdByNumber(client, &mockRESTClient{login: "testuser"}, repositoryFunc(t, orgRepository("myorg"), &called), 2)

		if !found || id != "PVT_user" {
			t.Errorf("got (%q, %v), want (PVT_user, true)", id, found)
		}
		if called {
			t.Error("repository should not be fetched")
		}
		assertQueryCall(t, client, 0, "UserProjectV2")
		assertVariable(t, client.queryCalls[0].Variables, "login", graphql.String("testuser"))
		assertVariable(t, client.queryCalls[0].Variables, "number", graphql.Int(2))
		if len(client.queryCalls) != 1 {
			t.Errorf("expected 1 query call, got %d", len(client.queryCalls))
		}
	})

	t.Run("ユーザーに無ければ組織のプロジェクトから探す", func(t *testing.T) {
		client := newClient("", "PVT_org")
		var called bool

		id, found := resolveProjectIdByNumber(client, &mockRESTClient{login: "testuser"}, repositoryFunc(t, orgRepository("myorg"), &called), 3)

		if !found || id != "PVT_org" {
			t.Errorf("got (%q, %v), want (PVT_org, true)", id, found)
		}
		assertQueryCall(t, client, 1, "OrgProjectV2")
		assertVariable(t, client.queryCalls[1].Variables, "login", graphql.String("myorg"))
		assertVariable(t, client.queryCalls[1].Variables, "number", graphql.Int(3))
	})

	t.Run("組織にも無ければ見つからない", func(t *testing.T) {
		client := newClient("", "")
		var called bool

		id, found := resolveProjectIdByNumber(client, &mockRESTClient{login: "testuser"}, repositoryFunc(t, orgRepository("myorg"), &called), 999)

		if found || id != "" {
			t.Errorf("got (%q, %v), want (\"\", false)", id, found)
		}
		if len(client.queryCalls) != 2 {
			t.Errorf("expected 2 query calls, got %d", len(client.queryCalls))
		}
	})

	t.Run("個人のリポジトリでユーザーに無ければ組織は探さず見つからない", func(t *testing.T) {
		client := newClient("", "PVT_org")
		var called bool

		_, found := resolveProjectIdByNumber(client, &mockRESTClient{login: "testuser"}, repositoryFunc(t, Repository{}, &called), 2)

		if found {
			t.Error("expected found to be false")
		}
		if len(client.queryCalls) != 1 {
			t.Errorf("expected 1 query call, got %d", len(client.queryCalls))
		}
	})
}

var mutationValueKindPattern = regexp.MustCompile(`value: \{(\w+):`)

// mutationValueKind は update mutation の graphql タグから値の種類(text / date / number / singleSelectOptionId / iterationId)を取り出す
func mutationValueKind(mutation interface{}) string {
	tag := reflect.TypeOf(mutation).Elem().Field(0).Tag.Get("graphql")
	m := mutationValueKindPattern.FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	return m[1]
}

type recordingGQLClient struct {
	*mockGQLClient
	kinds []string
}

func newRecordingGQLClient() *recordingGQLClient {
	c := &recordingGQLClient{mockGQLClient: &mockGQLClient{}}
	c.mutateFunc = func(name string, mutation interface{}, variables map[string]interface{}) error {
		c.kinds = append(c.kinds, mutationValueKind(mutation))
		return nil
	}
	return c
}

// failPrompter は呼ばれたらテストを失敗させる(CLI で値を指定したフィールドは対話しない)
func failPrompter(t *testing.T) fieldPrompter {
	return fieldPrompter{
		askText:   func(string) string { t.Error("askText should not be called"); return "" },
		askDate:   func(string) string { t.Error("askDate should not be called"); return "" },
		askNumber: func(string) float64 { t.Error("askNumber should not be called"); return 0 },
		askSelect: func(string, []Option) string { t.Error("askSelect should not be called"); return "" },
	}
}

var applyFieldValuesFields = []ProjectField{
	{Id: "f-text", Name: "Note", DataType: "TEXT"},
	{Id: "f-date", Name: "Due", DataType: "DATE"},
	{Id: "f-number", Name: "Point", DataType: "NUMBER"},
	{Id: "f-select", Name: "Status", DataType: "SINGLE_SELECT", Options: []Option{{Id: "Skip", Name: "Skip This Question."}, {Id: "opt-done", Name: "Done"}}},
	{Id: "f-iter", Name: "Iteration", DataType: "ITERATION", Options: []Option{{Id: "Skip", Name: "Skip This Question."}, {Id: "iter-1", Name: "2026-10-01"}}},
}

func TestApplyFieldValues(t *testing.T) {
	t.Run("CLI の値から型ごとの mutation を呼ぶ", func(t *testing.T) {
		tests := []struct {
			name      string
			field     string
			value     string
			wantKind  string
			wantField string
			wantValue interface{}
		}{
			{name: "TEXT", field: "Note", value: "hello", wantKind: "text", wantField: "f-text", wantValue: graphql.String("hello")},
			{name: "DATE", field: "Due", value: "2026-10-21", wantKind: "date", wantField: "f-date"},
			{name: "NUMBER", field: "Point", value: "3.5", wantKind: "number", wantField: "f-number", wantValue: graphql.Float(3.5)},
			{name: "SINGLE_SELECT", field: "Status", value: "Done", wantKind: "singleSelectOptionId", wantField: "f-select", wantValue: graphql.String("opt-done")},
			{name: "ITERATION", field: "Iteration", value: "2026-10-01", wantKind: "iterationId", wantField: "f-iter", wantValue: graphql.String("iter-1")},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				client := newRecordingGQLClient()
				var fields []ProjectField
				for _, f := range applyFieldValuesFields {
					if f.Name == tt.field {
						fields = append(fields, f)
					}
				}

				err := applyFieldValues(client, "proj-1", "item-1", fields, map[string]string{tt.field: tt.value}, failPrompter(t))

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(client.kinds, []string{tt.wantKind}) {
					t.Fatalf("mutation kinds: got %v, want [%s]", client.kinds, tt.wantKind)
				}
				vars := client.mutateCalls[0].Variables
				assertVariable(t, vars, "projectId", graphql.ID("proj-1"))
				assertVariable(t, vars, "itemId", graphql.ID("item-1"))
				assertVariable(t, vars, "fieldId", graphql.ID(tt.wantField))
				if tt.wantValue != nil {
					assertVariable(t, vars, "value", tt.wantValue)
				}
			})
		}
	})

	t.Run("全フィールドを CLI で指定すると定義順に全て更新する", func(t *testing.T) {
		client := newRecordingGQLClient()
		values := map[string]string{"Note": "n", "Due": "2026-10-21", "Point": "1", "Status": "Done", "Iteration": "2026-10-01"}

		err := applyFieldValues(client, "proj-1", "item-1", applyFieldValuesFields, values, failPrompter(t))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"text", "date", "number", "singleSelectOptionId", "iterationId"}
		if !reflect.DeepEqual(client.kinds, want) {
			t.Errorf("mutation kinds: got %v, want %v", client.kinds, want)
		}
	})

	t.Run("NUMBER に数値でない値はエラー", func(t *testing.T) {
		client := newRecordingGQLClient()
		fields := []ProjectField{applyFieldValuesFields[2]}

		err := applyFieldValues(client, "proj-1", "item-1", fields, map[string]string{"Point": "abc"}, failPrompter(t))

		var target invalidNumberError
		if !errors.As(err, &target) {
			t.Fatalf("expected invalidNumberError, got %v", err)
		}
		if err.Error() != "Invalid number value for field Point: abc" {
			t.Errorf("error message: got %q", err.Error())
		}
		if len(client.mutateCalls) != 0 {
			t.Errorf("expected no mutation, got %d", len(client.mutateCalls))
		}
	})

	t.Run("存在しない option 名はエラー", func(t *testing.T) {
		tests := []struct {
			name    string
			field   ProjectField
			wantMsg string
		}{
			{name: "SINGLE_SELECT", field: applyFieldValuesFields[3], wantMsg: "Option 'Unknown' not found for field 'Status'"},
			{name: "ITERATION", field: applyFieldValuesFields[4], wantMsg: "Option 'Unknown' not found for field 'Iteration'"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				client := newRecordingGQLClient()

				err := applyFieldValues(client, "proj-1", "item-1", []ProjectField{tt.field}, map[string]string{tt.field.Name: "Unknown"}, failPrompter(t))

				var target optionNotFoundError
				if !errors.As(err, &target) {
					t.Fatalf("expected optionNotFoundError, got %v", err)
				}
				if err.Error() != tt.wantMsg {
					t.Errorf("error message: got %q, want %q", err.Error(), tt.wantMsg)
				}
				if len(client.mutateCalls) != 0 {
					t.Errorf("expected no mutation, got %d", len(client.mutateCalls))
				}
			})
		}
	})

	t.Run("エラーになったフィールドより後は更新しない", func(t *testing.T) {
		client := newRecordingGQLClient()
		values := map[string]string{"Note": "n", "Due": "2026-10-21", "Point": "abc", "Status": "Done", "Iteration": "2026-10-01"}

		err := applyFieldValues(client, "proj-1", "item-1", applyFieldValuesFields, values, failPrompter(t))

		if err == nil {
			t.Fatal("expected error")
		}
		if !reflect.DeepEqual(client.kinds, []string{"text", "date"}) {
			t.Errorf("mutation kinds: got %v, want [text date]", client.kinds)
		}
	})

	t.Run("TEXT / DATE の空文字は mutation を呼ばない", func(t *testing.T) {
		client := newRecordingGQLClient()

		err := applyFieldValues(client, "proj-1", "item-1", applyFieldValuesFields[:2], map[string]string{"Note": "", "Due": ""}, failPrompter(t))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(client.mutateCalls) != 0 {
			t.Errorf("expected no mutation, got %d", len(client.mutateCalls))
		}
	})

	t.Run("CLI で指定しなかったフィールドは対話で尋ねる", func(t *testing.T) {
		client := newRecordingGQLClient()
		var asked []string
		prompter := fieldPrompter{
			askText:   func(name string) string { asked = append(asked, name); return "from prompt" },
			askDate:   func(name string) string { asked = append(asked, name); return "2026-10-21" },
			askNumber: func(name string) float64 { asked = append(asked, name); return 2 },
			askSelect: func(name string, options []Option) string {
				asked = append(asked, name)
				return options[1].Id
			},
		}

		err := applyFieldValues(client, "proj-1", "item-1", applyFieldValuesFields, map[string]string{}, prompter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(asked, []string{"Note", "Due", "Point", "Status", "Iteration"}) {
			t.Errorf("asked: got %v", asked)
		}
		want := []string{"text", "date", "number", "singleSelectOptionId", "iterationId"}
		if !reflect.DeepEqual(client.kinds, want) {
			t.Fatalf("mutation kinds: got %v, want %v", client.kinds, want)
		}
		assertVariable(t, client.mutateCalls[0].Variables, "value", graphql.String("from prompt"))
		assertVariable(t, client.mutateCalls[2].Variables, "value", graphql.Float(2))
		assertVariable(t, client.mutateCalls[3].Variables, "value", graphql.String("opt-done"))
		assertVariable(t, client.mutateCalls[4].Variables, "value", graphql.String("iter-1"))
	})

	t.Run("対話で空文字・Skip を選んだフィールドは更新しない", func(t *testing.T) {
		client := newRecordingGQLClient()
		prompter := fieldPrompter{
			askText:   func(string) string { return "" },
			askDate:   func(string) string { return "" },
			askNumber: func(string) float64 { t.Error("askNumber should not be called"); return 0 },
			askSelect: func(string, []Option) string { return "Skip" },
		}
		fields := []ProjectField{applyFieldValuesFields[0], applyFieldValuesFields[1], applyFieldValuesFields[3], applyFieldValuesFields[4]}

		err := applyFieldValues(client, "proj-1", "item-1", fields, map[string]string{}, prompter)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(client.mutateCalls) != 0 {
			t.Errorf("expected no mutation, got %d", len(client.mutateCalls))
		}
	})
}
