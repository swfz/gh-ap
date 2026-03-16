package main

import (
	"testing"
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
