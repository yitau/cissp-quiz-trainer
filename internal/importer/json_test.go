package importer

import (
	"encoding/json"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"os"
	"strings"
	"testing"
)

func TestValidation(t *testing.T) {
	data, err := os.ReadFile("../../samples/demo-questions.json")
	if err != nil {
		t.Fatal(err)
	}
	original, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*domain.QuestionFile)
	}{
		{"version", func(f *domain.QuestionFile) { f.SchemaVersion = "2.0" }},
		{"title", func(f *domain.QuestionFile) { f.Set.Title = " " }},
		{"empty", func(f *domain.QuestionFile) { f.Questions = nil }},
		{"duplicate", func(f *domain.QuestionFile) { f.Questions[1].ID = f.Questions[0].ID }},
		{"option", func(f *domain.QuestionFile) { delete(f.Questions[0].Options, "A") }},
		{"answer", func(f *domain.QuestionFile) { f.Questions[0].Answer = "E" }},
		{"explanation", func(f *domain.QuestionFile) { f.Questions[0].Explanation = " " }},
		{"domain", func(f *domain.QuestionFile) { f.Questions[0].Domain = 9 }},
		{"type", func(f *domain.QuestionFile) { f.Questions[0].Type = "UNKNOWN" }},
		{"difficulty", func(f *domain.QuestionFile) { f.Questions[0].Difficulty = "unknown" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var f domain.QuestionFile
			if err := json.Unmarshal(data, &f); err != nil {
				t.Fatal(err)
			}
			tc.change(&f)
			b, _ := json.Marshal(f)
			if _, err := Parse(b); err == nil {
				t.Fatal("accepted invalid file")
			}
		})
	}
	for _, b := range []string{"{", string(data) + "{}", strings.Replace(string(data), `"schemaVersion": "1.0"`, `"schemaVersion":"1.0","schemaVersion":"1.0"`, 1), strings.Replace(string(data), `"domain":1`, `"domain":"1"`, 1)} {
		if _, err := Parse([]byte(b)); err == nil {
			t.Fatal("accepted malformed JSON")
		}
	}
	if len(original.Questions) != 4 {
		t.Fatal("sample count")
	}
}
