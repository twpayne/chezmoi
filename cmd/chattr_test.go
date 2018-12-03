package cmd

import (
	"reflect"
	"testing"
)

func TestParseAttributeModifiers(t *testing.T) {
	for _, tc := range []struct {
		s       string
		want    *attributeModifiers
		wantErr bool
	}{
		{s: "empty", want: &attributeModifiers{empty: 1}},
		{s: "+empty", want: &attributeModifiers{empty: 1}},
		{s: "-empty", want: &attributeModifiers{empty: -1}},
		{s: "e", want: &attributeModifiers{empty: 1}},
		{s: "+e", want: &attributeModifiers{empty: 1}},
		{s: "-e", want: &attributeModifiers{empty: -1}},
		{s: "executable", want: &attributeModifiers{executable: 1}},
		{s: "+executable", want: &attributeModifiers{executable: 1}},
		{s: "-executable", want: &attributeModifiers{executable: -1}},
		{s: "x", want: &attributeModifiers{executable: 1}},
		{s: "+x", want: &attributeModifiers{executable: 1}},
		{s: "-x", want: &attributeModifiers{executable: -1}},
		{s: "private", want: &attributeModifiers{private: 1}},
		{s: "+private", want: &attributeModifiers{private: 1}},
		{s: "-private", want: &attributeModifiers{private: -1}},
		{s: "p", want: &attributeModifiers{private: 1}},
		{s: "+p", want: &attributeModifiers{private: 1}},
		{s: "-p", want: &attributeModifiers{private: -1}},
		{s: "template", want: &attributeModifiers{template: 1}},
		{s: "+template", want: &attributeModifiers{template: 1}},
		{s: "-template", want: &attributeModifiers{template: -1}},
		{s: "t", want: &attributeModifiers{template: 1}},
		{s: "+t", want: &attributeModifiers{template: 1}},
		{s: "-t", want: &attributeModifiers{template: -1}},
		{s: "empty,executable,private,template", want: &attributeModifiers{empty: 1, executable: 1, private: 1, template: 1}},
		{s: "+empty,+executable,+private,+template", want: &attributeModifiers{empty: 1, executable: 1, private: 1, template: 1}},
		{s: "-empty,-executable,-private,-template", want: &attributeModifiers{empty: -1, executable: -1, private: -1, template: -1}},
		{s: "foo", wantErr: true},
		{s: "empty,foo", wantErr: true},
		{s: "empty,foo", wantErr: true},
		{s: " empty , -private ", want: &attributeModifiers{empty: 1, private: -1}},
		{s: "empty,,-private", want: &attributeModifiers{empty: 1, private: -1}},
	} {
		if got, gotErr := parseAttributeModifiers(tc.s); (gotErr != nil && !tc.wantErr) || (gotErr == nil && tc.wantErr) || !reflect.DeepEqual(got, tc.want) {
			wantErrStr := "<nil>"
			if tc.wantErr {
				wantErrStr = "!" + wantErrStr
			}
			t.Errorf("parseAttributeModifiers(%q) == %+v, %v, want %+v, %s", tc.s, got, gotErr, tc.want, wantErrStr)
		}
	}
}
