package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type gqlCall struct {
	Name      string
	Variables map[string]interface{}
}

type mockGQLClient struct {
	queryCalls  []gqlCall
	mutateCalls []gqlCall
	queryFunc   func(name string, query interface{}, variables map[string]interface{}) error
	mutateFunc  func(name string, mutation interface{}, variables map[string]interface{}) error
}

func (m *mockGQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	return nil
}

func (m *mockGQLClient) DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	return nil
}

func (m *mockGQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	m.queryCalls = append(m.queryCalls, gqlCall{Name: name, Variables: variables})
	if m.queryFunc != nil {
		return m.queryFunc(name, query, variables)
	}
	return nil
}

func (m *mockGQLClient) QueryWithContext(ctx context.Context, name string, query interface{}, variables map[string]interface{}) error {
	return nil
}

func (m *mockGQLClient) Mutate(name string, mutation interface{}, variables map[string]interface{}) error {
	m.mutateCalls = append(m.mutateCalls, gqlCall{Name: name, Variables: variables})
	if m.mutateFunc != nil {
		return m.mutateFunc(name, mutation, variables)
	}
	return nil
}

func (m *mockGQLClient) MutateWithContext(ctx context.Context, name string, mutation interface{}, variables map[string]interface{}) error {
	return nil
}

// setNestedField はreflectを使ってGQLクエリ構造体のネストしたフィールドにデータをセットする。
// pathはドット区切りのフィールドパス（例: "Node.ProjectV2.Fields.Nodes"）。
// valueが直接代入できない型（匿名構造体のスライス等）の場合、フィールドごとにコピーする。
func setNestedField(query interface{}, path string, value interface{}) {
	v := reflect.ValueOf(query).Elem()
	for _, field := range splitPath(path) {
		v = v.FieldByName(field)
		if !v.IsValid() {
			panic(fmt.Sprintf("field %q not found in path %q", field, path))
		}
	}

	srcVal := reflect.ValueOf(value)

	// 型が一致すればそのままセット
	if srcVal.Type().AssignableTo(v.Type()) {
		v.Set(srcVal)
		return
	}

	// スライスの場合、要素ごとにフィールドをコピー
	if v.Kind() == reflect.Slice && srcVal.Kind() == reflect.Slice {
		destSlice := reflect.MakeSlice(v.Type(), srcVal.Len(), srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			copyStructFields(destSlice.Index(i), srcVal.Index(i))
		}
		v.Set(destSlice)
		return
	}

	// 構造体の場合、フィールドをコピー
	if v.Kind() == reflect.Struct && srcVal.Kind() == reflect.Struct {
		copyStructFields(v, srcVal)
		return
	}

	panic(fmt.Sprintf("cannot assign %v to %v", srcVal.Type(), v.Type()))
}

// copyStructFields はフィールド名が一致するもの同士をコピーする
func copyStructFields(dest, src reflect.Value) {
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		fieldName := srcType.Field(i).Name
		destField := dest.FieldByName(fieldName)
		if destField.IsValid() && destField.CanSet() {
			srcField := src.Field(i)
			if srcField.Type().AssignableTo(destField.Type()) {
				destField.Set(srcField)
			} else if destField.Kind() == reflect.Slice && srcField.Kind() == reflect.Slice {
				innerSlice := reflect.MakeSlice(destField.Type(), srcField.Len(), srcField.Len())
				for j := 0; j < srcField.Len(); j++ {
					copyStructFields(innerSlice.Index(j), srcField.Index(j))
				}
				destField.Set(innerSlice)
			} else if destField.Kind() == reflect.Struct && srcField.Kind() == reflect.Struct {
				copyStructFields(destField, srcField)
			}
		}
	}
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func assertQueryCall(t *testing.T, client *mockGQLClient, index int, wantName string) {
	t.Helper()
	if len(client.queryCalls) <= index {
		t.Fatalf("expected at least %d query calls, got %d", index+1, len(client.queryCalls))
	}
	if client.queryCalls[index].Name != wantName {
		t.Errorf("query call[%d] name: got %q, want %q", index, client.queryCalls[index].Name, wantName)
	}
}

func assertMutateCall(t *testing.T, client *mockGQLClient, index int, wantName string) {
	t.Helper()
	if len(client.mutateCalls) <= index {
		t.Fatalf("expected at least %d mutate calls, got %d", index+1, len(client.mutateCalls))
	}
	if client.mutateCalls[index].Name != wantName {
		t.Errorf("mutate call[%d] name: got %q, want %q", index, client.mutateCalls[index].Name, wantName)
	}
}

func assertVariable(t *testing.T, variables map[string]interface{}, key string, want interface{}) {
	t.Helper()
	got, ok := variables[key]
	if !ok {
		t.Errorf("variable %q not found", key)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("variable %q: got %v (%T), want %v (%T)", key, got, got, want, want)
	}
}
