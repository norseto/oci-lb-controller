/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

For the full license text, please visit: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testSpec struct {
	Foo string `json:"foo"`
	Num int    `json:"num"`
}

func TestSaveSpecInAnnotationsPreservesUnrelated(t *testing.T) {
	meta := &metav1.ObjectMeta{
		Annotations: map[string]string{
			"existing": "value",
		},
	}
	spec := testSpec{Foo: "bar", Num: 1}

	if err := saveSpecInAnnotations(meta, spec); err != nil {
		t.Fatalf("saveSpecInAnnotations returned error: %v", err)
	}

	if meta.Annotations["existing"] != "value" {
		t.Errorf("existing annotation changed: %v", meta.Annotations["existing"])
	}
	if _, ok := meta.Annotations[specAnnotation]; !ok {
		t.Errorf("spec annotation not set")
	}
	if len(meta.Annotations) != 2 {
		t.Errorf("annotation count mismatch: expected 2 got %d", len(meta.Annotations))
	}
}

func TestGetSpecInAnnotationsRetrievesSavedSpec(t *testing.T) {
	meta := &metav1.ObjectMeta{}
	expected := testSpec{Foo: "baz", Num: 99}
	if err := saveSpecInAnnotations(meta, expected); err != nil {
		t.Fatalf("saving spec failed: %v", err)
	}
	var actual testSpec
	ok, err := getSpecInAnnotations(meta, &actual)
	if err != nil {
		t.Fatalf("getSpecInAnnotations returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected annotation to be present")
	}
	if actual != expected {
		t.Errorf("retrieved spec mismatch: expected %#v got %#v", expected, actual)
	}
}

func TestHasSpecAnnotation(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        bool
	}{
		{
			name:        "no annotations",
			annotations: nil,
			want:        false,
		},
		{
			name:        "annotations without spec",
			annotations: map[string]string{"other": "value"},
			want:        false,
		},
		{
			name:        "spec annotation with empty value",
			annotations: map[string]string{specAnnotation: ""},
			want:        false,
		},
		{
			name:        "spec annotation with non-empty value",
			annotations: map[string]string{specAnnotation: `{"foo":"bar"}`},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := &metav1.ObjectMeta{Annotations: tt.annotations}
			if got := hasSpecAnnotation(meta); got != tt.want {
				t.Errorf("hasSpecAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}
