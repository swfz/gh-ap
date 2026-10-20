package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type fatalCalled struct {
	message string
}

// fakeGh は ghExec / logFatal を差し替え、渡された引数を記録して stdout / err を返す。
// logFatal は終了せずに panic(fatalCalled) するので、呼び出し側で recover して検証する。
func fakeGh(t *testing.T, stdout string, execErr error) *[][]string {
	t.Helper()
	var calls [][]string
	origExec, origFatal := ghExec, logFatal
	t.Cleanup(func() {
		ghExec, logFatal = origExec, origFatal
	})
	ghExec = func(args ...string) (bytes.Buffer, bytes.Buffer, error) {
		calls = append(calls, args)
		return *bytes.NewBufferString(stdout), bytes.Buffer{}, execErr
	}
	logFatal = func(v ...interface{}) {
		panic(fatalCalled{message: fmt.Sprint(v...)})
	}
	return &calls
}

// recoverValue は f を実行し、panic した値を返す(panic しなければ nil)
func recoverValue(f func()) (v interface{}) {
	defer func() {
		v = recover()
	}()
	f()
	return nil
}

func assertArgs(t *testing.T, calls *[][]string, want []string) {
	t.Helper()
	if len(*calls) != 1 {
		t.Fatalf("expected 1 gh call, got %d", len(*calls))
	}
	if !reflect.DeepEqual((*calls)[0], want) {
		t.Errorf("gh args: got %v, want %v", (*calls)[0], want)
	}
}

func assertFatal(t *testing.T, v interface{}, wantMessage string) {
	t.Helper()
	f, ok := v.(fatalCalled)
	if !ok {
		t.Fatalf("expected logFatal to be called, got %v", v)
	}
	if f.message != wantMessage {
		t.Errorf("fatal message: got %q, want %q", f.message, wantMessage)
	}
}

func assertJSONPanic(t *testing.T, v interface{}) {
	t.Helper()
	if _, ok := v.(error); !ok {
		t.Fatalf("expected panic with JSON error, got %v", v)
	}
}

func TestGhRepository(t *testing.T) {
	t.Run("JSON をパースする", func(t *testing.T) {
		calls := fakeGh(t, `{"name":"gh-ap","owner":{"id":"O_1","login":"swfz"},"isInOrganization":true}`, nil)

		repository := ghRepository()

		assertArgs(t, calls, []string{"repo", "view", "--json", "name,owner,isInOrganization"})
		want := orgRepositoryWithId("gh-ap", "O_1", "swfz")
		if !reflect.DeepEqual(repository, want) {
			t.Errorf("got %+v, want %+v", repository, want)
		}
	})

	t.Run("実行エラーは logFatal", func(t *testing.T) {
		fakeGh(t, "", errors.New("not a git repository"))

		assertFatal(t, recoverValue(func() { ghRepository() }), "not a git repository")
	})

	t.Run("不正な JSON は panic", func(t *testing.T) {
		fakeGh(t, "not json", nil)

		assertJSONPanic(t, recoverValue(func() { ghRepository() }))
	})
}

func orgRepositoryWithId(name, ownerId, ownerLogin string) Repository {
	r := Repository{Name: name, IsInOrganization: true}
	r.Owner.Id = ownerId
	r.Owner.Login = ownerLogin
	return r
}

func TestGhCurrentPullRequest(t *testing.T) {
	t.Run("JSON をパースする", func(t *testing.T) {
		calls := fakeGh(t, `{"id":"PR_1","number":12,"title":"current"}`, nil)

		content := ghCurrentPullRequest()

		assertArgs(t, calls, []string{"pr", "view", "--json", "id,number,title"})
		want := Content{Id: "PR_1", Number: 12, Title: "current"}
		if !reflect.DeepEqual(content, want) {
			t.Errorf("got %+v, want %+v", content, want)
		}
	})

	t.Run("実行エラーは logFatal", func(t *testing.T) {
		fakeGh(t, "", errors.New("no pull requests found"))

		assertFatal(t, recoverValue(func() { ghCurrentPullRequest() }), "no pull requests found")
	})

	t.Run("不正な JSON は panic", func(t *testing.T) {
		fakeGh(t, "{", nil)

		assertJSONPanic(t, recoverValue(func() { ghCurrentPullRequest() }))
	})
}

func TestGhContent(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantArgs    []string
	}{
		{name: "Issue は issue view", contentType: "Issue", wantArgs: []string{"issue", "view", "5", "--json", "id,number,title"}},
		{name: "PullRequest は pr view", contentType: "PullRequest", wantArgs: []string{"pr", "view", "5", "--json", "id,number,title"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := fakeGh(t, `{"id":"I_5","number":5,"title":"five"}`, nil)

			content := ghContent(tt.contentType, "5")

			assertArgs(t, calls, tt.wantArgs)
			want := Content{Id: "I_5", Number: 5, Title: "five"}
			if !reflect.DeepEqual(content, want) {
				t.Errorf("got %+v, want %+v", content, want)
			}
		})
	}

	t.Run("実行エラーは種類つきのメッセージで logFatal", func(t *testing.T) {
		fakeGh(t, "", errors.New("exit status 1"))

		assertFatal(t, recoverValue(func() { ghContent("Issue", "999") }), "Error: not found Issue")
	})

	t.Run("不正な JSON は panic", func(t *testing.T) {
		fakeGh(t, "[]", nil)

		assertJSONPanic(t, recoverValue(func() { ghContent("PullRequest", "1") }))
	})
}

func TestGhContentList(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantArgs    []string
	}{
		{name: "Issue は issue list", contentType: "Issue", wantArgs: []string{"issue", "list", "--limit", "50", "--json", "id,number,title"}},
		{name: "PullRequest は pr list", contentType: "PullRequest", wantArgs: []string{"pr", "list", "--limit", "50", "--json", "id,number,title"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := fakeGh(t, `[{"id":"I_1","number":1,"title":"one"},{"id":"I_2","number":2,"title":"two"}]`, nil)

			contents := ghContentList(tt.contentType)

			assertArgs(t, calls, tt.wantArgs)
			want := []Content{{Id: "I_1", Number: 1, Title: "one"}, {Id: "I_2", Number: 2, Title: "two"}}
			if !reflect.DeepEqual(contents, want) {
				t.Errorf("got %+v, want %+v", contents, want)
			}
		})
	}

	t.Run("実行エラーは logFatal", func(t *testing.T) {
		fakeGh(t, "", errors.New("exit status 1"))

		assertFatal(t, recoverValue(func() { ghContentList("Issue") }), "exit status 1")
	})

	t.Run("不正な JSON は panic", func(t *testing.T) {
		fakeGh(t, "{}", nil)

		assertJSONPanic(t, recoverValue(func() { ghContentList("Issue") }))
	})
}
