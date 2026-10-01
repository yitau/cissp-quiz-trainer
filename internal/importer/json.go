package importer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yitau/cissp-quiz-trainer/internal/domain"
)

const MaxFileBytes = 16 << 20

func Parse(data []byte) (domain.QuestionFile, error) {
	var f domain.QuestionFile
	if len(data) > MaxFileBytes {
		return f, fmt.Errorf("题集超过 16 MiB 限制")
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	d := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueKeys(d); err != nil {
		return f, fmt.Errorf("JSON 格式错误：%w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return f, fmt.Errorf("JSON 根对象之后存在额外内容")
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return f, fmt.Errorf("字段类型错误：%w", err)
	}
	if f.SchemaVersion != "1.0" {
		return f, fmt.Errorf("schemaVersion 必须为 1.0，收到 %q", f.SchemaVersion)
	}
	if blank(f.Set.ID) || blank(f.Set.Title) {
		return f, fmt.Errorf("set.id 和 set.title 不能为空")
	}
	if f.Set.Domain != 0 && (f.Set.Domain < 1 || f.Set.Domain > 8) {
		return f, fmt.Errorf("set.domain 必须为 1-8")
	}
	if f.Set.Difficulty != "" && !oneOf(f.Set.Difficulty, "easy", "medium", "hard") {
		return f, fmt.Errorf("set.difficulty 无效")
	}
	if len(f.Questions) == 0 {
		return f, fmt.Errorf("questions 至少需要一道题")
	}
	seen := map[string]bool{}
	for i, q := range f.Questions {
		prefix := fmt.Sprintf("questions[%d] (%s)", i, q.ID)
		if blank(q.ID) || blank(q.Question) || blank(q.Explanation) {
			return f, fmt.Errorf("%s：id/question/explanation 不能为空", prefix)
		}
		if seen[q.ID] {
			return f, fmt.Errorf("%s：重复题目 ID", prefix)
		}
		seen[q.ID] = true
		if len(q.Options) != 4 {
			return f, fmt.Errorf("%s：options 必须且只能包含 A/B/C/D", prefix)
		}
		for _, k := range []string{"A", "B", "C", "D"} {
			if blank(q.Options[k]) {
				return f, fmt.Errorf("%s：options.%s 不能为空", prefix, k)
			}
		}
		if !oneOf(q.Answer, "A", "B", "C", "D") {
			return f, fmt.Errorf("%s：answer 必须为 A/B/C/D", prefix)
		}
		if q.Domain < 1 || q.Domain > 8 {
			return f, fmt.Errorf("%s：domain 必须为 1-8", prefix)
		}
		if !oneOf(q.Type, "FIRST", "BEST", "MOST", "NEXT", "PRIMARY", "LEAST", "MOST_LIKELY", "SCENARIO", "KNOWLEDGE") {
			return f, fmt.Errorf("%s：type 不在支持的枚举内", prefix)
		}
		if !oneOf(q.Difficulty, "easy", "medium", "hard") {
			return f, fmt.Errorf("%s：difficulty 必须为 easy/medium/hard", prefix)
		}
		for k, v := range q.WhyOthersWrong {
			if !oneOf(k, "A", "B", "C", "D") || k == q.Answer || blank(v) {
				return f, fmt.Errorf("%s：whyOthersWrong.%s 无效", prefix, k)
			}
		}
		for _, tag := range q.Tags {
			if blank(tag) {
				return f, fmt.Errorf("%s：tags 不能包含空标签", prefix)
			}
		}
	}
	return f, nil
}
func blank(s string) bool { return strings.TrimSpace(s) == "" }
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

// encoding/json alone accepts duplicate properties; reject ambiguous input first.
func uniqueKeys(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			k, ok := t.(string)
			if !ok {
				return fmt.Errorf("对象属性必须为字符串")
			}
			if seen[k] {
				return fmt.Errorf("重复属性 %q", k)
			}
			seen[k] = true
			if err := uniqueKeys(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueKeys(d); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("意外分隔符")
	}
	_, err = d.Token()
	return err
}
