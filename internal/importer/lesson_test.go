package importer

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func lessonSample(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../samples/cissp-week01-day01-lesson.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestLessonSample(t *testing.T) {
	f, e := ParseLesson(lessonSample(t))
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Sections) != 2 || len(f.Concepts) != 8 || f.CaseStudy == nil || *f.Lesson.LinkedQuestionSetID != "w01d01-review-20261008-v1" {
		t.Fatalf("unexpected sample: %+v", f.Lesson)
	}
}

func TestLessonOptionalAndCanonicalNumbers(t *testing.T) {
	var root map[string]any
	json.Unmarshal(lessonSample(t), &root)
	l := root["lesson"].(map[string]any)
	delete(l, "description")
	l["linkedQuestionSetId"] = nil
	delete(root, "caseStudy")
	for _, item := range root["concepts"].([]any) {
		delete(item.(map[string]any), "controls")
	}
	b, _ := json.Marshal(root)
	b = bytes.Replace(b, []byte(`"week":1`), []byte(`"week":1e0`), 1)
	f, err := ParseLesson(b)
	if err != nil || f.Lesson.Week != 1 || f.CaseStudy != nil || f.Lesson.LinkedQuestionSetID != nil {
		t.Fatal(f.Lesson, err)
	}
	if _, err := Parse(b); err == nil || !strings.Contains(err.Error(), "知识学习") {
		t.Fatal("lesson accepted by question importer", err)
	}
	if _, err := ParseLesson([]byte(`{"set":{},"schemaVersion":"1.0","questions":[]}`)); err == nil || !strings.Contains(err.Error(), "题库与导入") {
		t.Fatal(err)
	}
}
func TestLessonValidation(t *testing.T) {
	cases := []struct {
		name, path string
		edit       func(map[string]any)
	}{
		{"format", "format", func(m map[string]any) { m["format"] = "other" }},
		{"version", "schemaVersion", func(m map[string]any) { m["schemaVersion"] = "2.0" }},
		{"missing", "lesson.title", func(m map[string]any) { delete(m["lesson"].(map[string]any), "title") }},
		{"blank", "lesson.title", func(m map[string]any) { m["lesson"].(map[string]any)["title"] = " \t" }},
		{"null", "lesson.title", func(m map[string]any) { m["lesson"].(map[string]any)["title"] = nil }},
		{"type", "lesson.week", func(m map[string]any) { m["lesson"].(map[string]any)["week"] = "1" }},
		{"week", "lesson.week", func(m map[string]any) { m["lesson"].(map[string]any)["week"] = 61 }},
		{"day", "lesson.day", func(m map[string]any) { m["lesson"].(map[string]any)["day"] = 0 }},
		{"domain", "lesson.domain", func(m map[string]any) { m["lesson"].(map[string]any)["domain"] = 9 }},
		{"date", "lesson.createdAt", func(m map[string]any) { m["lesson"].(map[string]any)["createdAt"] = "2026-10-08" }},
		{"unknown", "lesson.extra", func(m map[string]any) { m["lesson"].(map[string]any)["extra"] = true }},
		{"empty", "sections", func(m map[string]any) { m["sections"] = []any{} }},
		{"duplicate section", "sections[1].id", func(m map[string]any) {
			ss := m["sections"].([]any)
			ss[1].(map[string]any)["id"] = ss[0].(map[string]any)["id"]
		}},
		{"duplicate concept", "concepts[1].id", func(m map[string]any) {
			cs := m["concepts"].([]any)
			cs[1].(map[string]any)["id"] = cs[0].(map[string]any)["id"]
		}},
		{"unknown section", "sectionId", func(m map[string]any) { m["concepts"].([]any)[0].(map[string]any)["sectionId"] = "missing" }},
		{"wrong section", "conceptIds", func(m map[string]any) { m["concepts"].([]any)[0].(map[string]any)["sectionId"] = "risk-basics" }},
		{"missing reference", "conceptIds", func(m map[string]any) {
			m["sections"].([]any)[0].(map[string]any)["conceptIds"] = []any{"w01d01-integrity", "w01d01-availability"}
		}},
		{"unknown reference", "conceptIds", func(m map[string]any) { m["sections"].([]any)[0].(map[string]any)["conceptIds"] = []any{"missing"} }},
		{"duplicate reference", "conceptIds", func(m map[string]any) {
			m["sections"].([]any)[0].(map[string]any)["conceptIds"] = []any{"w01d01-integrity", "w01d01-integrity"}
		}},
		{"identifier", "lesson.id", func(m map[string]any) { m["lesson"].(map[string]any)["id"] = "bad id" }},
		{"long text", "lesson.title", func(m map[string]any) { m["lesson"].(map[string]any)["title"] = strings.Repeat("中", 20001) }},
		{"empty controls", "controls", func(m map[string]any) { m["concepts"].([]any)[0].(map[string]any)["controls"] = []any{} }},
		{"null case", "caseStudy", func(m map[string]any) { m["caseStudy"] = nil }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var m map[string]any
			json.Unmarshal(lessonSample(t), &m)
			tt.edit(m)
			b, _ := json.Marshal(m)
			_, e := ParseLesson(b)
			if e == nil || !strings.Contains(e.Error(), tt.path) {
				t.Fatalf("want path %s, got %v", tt.path, e)
			}
		})
	}
	for _, b := range [][]byte{[]byte(`{"format":"cissp-lesson","format":"cissp-lesson"}`), []byte(strings.Repeat(" ", MaxFileBytes+1)), append(lessonSample(t), []byte(` {}`)...)} {
		if _, e := ParseLesson(b); e == nil {
			t.Fatal("accepted invalid JSON")
		}
	}
}
