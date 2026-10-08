package importer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

// ParseLesson implements the closed Lesson v1 contract, independently of Question v1.
func ParseLesson(data []byte) (domain.LessonFile, error) {
	var f domain.LessonFile
	if len(data) > MaxFileBytes {
		return f, fmt.Errorf("课程超过 16 MiB 限制")
	}
	if !utf8.Valid(data) {
		return f, fmt.Errorf("课程 JSON 必须为 UTF-8")
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	d := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueKeys(d); err != nil {
		return f, fmt.Errorf("JSON 格式错误：%w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return f, fmt.Errorf("JSON 根对象之后存在额外内容")
	}
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return f, err
	}
	if m, ok := root.(map[string]any); ok && m["set"] != nil {
		return f, fmt.Errorf("文件是题库 JSON，请在题库与导入中打开")
	}
	v := lessonValidator{}
	m := v.object(root, "$", "format schemaVersion lesson sections concepts", "caseStudy")
	if v.text(m, "format", "format") != "cissp-lesson" {
		v.fail("format", "必须为 cissp-lesson")
	}
	if v.text(m, "schemaVersion", "schemaVersion") != "1.0" {
		v.fail("schemaVersion", "必须为 1.0")
	}
	l := v.object(m["lesson"], "lesson", "id title week day domain createdAt", "description linkedQuestionSetId")
	v.id(l, "id", "lesson.id")
	v.text(l, "title", "lesson.title")
	if _, ok := l["description"]; ok {
		v.text(l, "description", "lesson.description")
	}
	v.integer(l, "week", 60)
	v.integer(l, "day", 7)
	v.integer(l, "domain", 8)
	stamp := v.text(l, "createdAt", "lesson.createdAt")
	if _, err := time.Parse(time.RFC3339Nano, stamp); err != nil {
		v.fail("lesson.createdAt", "需要带时区的有效日期时间")
	}
	if value, ok := l["linkedQuestionSetId"]; ok && value != nil {
		v.id(l, "linkedQuestionSetId", "lesson.linkedQuestionSetId")
	}
	for i, item := range v.array(m["sections"], "sections", 30) {
		p := fmt.Sprintf("sections[%d]", i)
		s := v.object(item, p, "id title conceptIds", "")
		v.id(s, "id", p+".id")
		v.text(s, "title", p+".title")
		v.strings(s["conceptIds"], p+".conceptIds", 200, true)
	}
	for i, item := range v.array(m["concepts"], "concepts", 200) {
		p := fmt.Sprintf("concepts[%d]", i)
		c := v.object(item, p, "id sectionId term objective definition explanation analogy scenario commonMistake examTip recall", "controls")
		v.id(c, "id", p+".id")
		v.id(c, "sectionId", p+".sectionId")
		for _, k := range strings.Fields("objective definition explanation analogy scenario commonMistake examTip") {
			v.text(c, k, p+"."+k)
		}
		term := v.object(c["term"], p+".term", "en zh", "")
		v.text(term, "en", p+".term.en")
		v.text(term, "zh", p+".term.zh")
		recall := v.object(c["recall"], p+".recall", "prompt answer", "")
		v.text(recall, "prompt", p+".recall.prompt")
		v.text(recall, "answer", p+".recall.answer")
		if controls, ok := c["controls"]; ok {
			v.strings(controls, p+".controls", 30, false)
		}
	}
	if item, ok := m["caseStudy"]; ok {
		c := v.object(item, "caseStudy", "title background steps takeaway", "")
		for _, k := range []string{"title", "background", "takeaway"} {
			v.text(c, k, "caseStudy."+k)
		}
		for i, item := range v.array(c["steps"], "caseStudy.steps", 30) {
			p := fmt.Sprintf("caseStudy.steps[%d]", i)
			s := v.object(item, p, "label detail", "")
			v.text(s, "label", p+".label")
			v.text(s, "detail", p+".detail")
		}
	}
	if v.err != nil {
		return f, v.err
	}
	// JSON Schema integers include numeric spellings such as 1.0 and 1e0.
	// Re-encoding the validated tree gives these the same canonical representation.
	normalized, err := json.Marshal(root)
	if err != nil {
		return f, err
	}
	if err := json.Unmarshal(normalized, &f); err != nil {
		return f, fmt.Errorf("字段类型错误：%w", err)
	}
	sections := map[string]bool{}
	concepts := map[string]domain.Concept{}
	refs := map[string]bool{}
	for i, s := range f.Sections {
		if sections[s.ID] {
			return f, fmt.Errorf("sections[%d].id (%s)：重复章节 ID", i, s.ID)
		}
		sections[s.ID] = true
	}
	for i, c := range f.Concepts {
		if _, ok := concepts[c.ID]; ok {
			return f, fmt.Errorf("concepts[%d].id (%s)：重复知识点 ID", i, c.ID)
		}
		if !sections[c.SectionID] {
			return f, fmt.Errorf("concepts[%d].sectionId (%s)：章节不存在", i, c.ID)
		}
		concepts[c.ID] = c
	}
	for i, s := range f.Sections {
		for j, id := range s.ConceptIDs {
			c, ok := concepts[id]
			if !ok || refs[id] || c.SectionID != s.ID {
				return f, fmt.Errorf("sections[%d].conceptIds[%d] (%s)：引用不存在、重复或章节不一致", i, j, id)
			}
			refs[id] = true
		}
	}
	for _, c := range f.Concepts {
		if !refs[c.ID] {
			return f, fmt.Errorf("sections[].conceptIds：遗漏知识点 %s", c.ID)
		}
	}
	return f, nil
}

// Small, lesson-specific checks keep optional nulls and absent fields distinct.
type lessonValidator struct{ err error }

func (v *lessonValidator) fail(path, message string) {
	if v.err == nil {
		v.err = fmt.Errorf("%s：%s", path, message)
	}
}
func (v *lessonValidator) object(value any, path, required, optional string) map[string]any {
	m, ok := value.(map[string]any)
	if !ok {
		v.fail(path, "必须为对象")
		return map[string]any{}
	}
	allowed := map[string]bool{}
	for _, k := range strings.Fields(required) {
		allowed[k] = true
		if _, ok := m[k]; !ok {
			v.fail(path+"."+k, "缺少必填字段")
		}
	}
	for _, k := range strings.Fields(optional) {
		allowed[k] = true
	}
	for k := range m {
		if !allowed[k] {
			v.fail(path+"."+k, "未知字段")
		}
	}
	return m
}
func (v *lessonValidator) text(m map[string]any, key, path string) string {
	s, ok := m[key].(string)
	if !ok || blank(s) || utf8.RuneCountInString(s) > 20000 {
		v.fail(path, "必须为 1–20000 字符的非空白文本")
	}
	return s
}
func (v *lessonValidator) id(m map[string]any, key, path string) {
	s := v.text(m, key, path)
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`).MatchString(s) {
		v.fail(path, "必须为 1–128 字符的英文、数字、点、下划线或连字符标识")
	}
}
func (v *lessonValidator) integer(m map[string]any, key string, max int) {
	n, ok := m[key].(float64)
	if !ok || n < 1 || n > float64(max) || n != float64(int(n)) {
		v.fail("lesson."+key, fmt.Sprintf("必须为 1–%d 的整数", max))
	}
}
func (v *lessonValidator) array(value any, path string, max int) []any {
	a, ok := value.([]any)
	if !ok || len(a) < 1 || len(a) > max {
		v.fail(path, fmt.Sprintf("必须为 1–%d 项的数组", max))
	}
	return a
}
func (v *lessonValidator) strings(value any, path string, max int, ids bool) {
	seen := map[string]bool{}
	for i, item := range v.array(value, path, max) {
		p := fmt.Sprintf("%s[%d]", path, i)
		m := map[string]any{"value": item}
		s := v.text(m, "value", p)
		if ids {
			v.id(m, "value", p)
		}
		if seen[s] {
			v.fail(p, "重复值")
		}
		seen[s] = true
	}
}
