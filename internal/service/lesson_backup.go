package service

import (
	"encoding/json"
	"fmt"
	"github.com/yitau/cissp-quiz-trainer/internal/domain"
	"github.com/yitau/cissp-quiz-trainer/internal/importer"
	"time"
)

func validateLessonContents(c domain.BackupContents) error {
	lessons := map[string]map[string]bool{}
	for _, s := range c.Lessons {
		f, err := importer.ParseLesson([]byte(s.Document))
		if err != nil {
			return fmt.Errorf("备份课程 %s：%w", s.ID, err)
		}
		b, _ := json.Marshal(f)
		if s.ID != f.Lesson.ID || s.Fingerprint != digest(b) || s.Document != string(b) {
			return fmt.Errorf("备份课程 %s 身份或指纹不一致", s.ID)
		}
		if _, ok := lessons[s.ID]; ok {
			return fmt.Errorf("备份课程 ID 重复")
		}
		if _, err := time.Parse(time.RFC3339Nano, s.ImportedAt); err != nil {
			return fmt.Errorf("课程导入时间无效")
		}
		concepts := map[string]bool{}
		for _, concept := range f.Concepts {
			concepts[concept.ID] = true
		}
		lessons[s.ID] = concepts
	}
	seen := map[[2]string]bool{}
	for _, p := range c.LessonProgress {
		if !lessons[p.UnitID][p.ConceptID] {
			return fmt.Errorf("learning_progress %s/%s 引用不存在", p.UnitID, p.ConceptID)
		}
		key := [2]string{p.UnitID, p.ConceptID}
		if seen[key] {
			return fmt.Errorf("learning_progress 重复记录")
		}
		seen[key] = true
		if p.Status != "in_progress" && p.Status != "understood" && p.Status != "needs_review" {
			return fmt.Errorf("learning_progress 状态无效")
		}
		updated, err := time.Parse(time.RFC3339Nano, p.UpdatedAt)
		if err != nil {
			return fmt.Errorf("learning_progress.updated_at 无效")
		}
		if p.LastOpenedAt != "" {
			opened, err := time.Parse(time.RFC3339Nano, p.LastOpenedAt)
			if err != nil || opened.After(updated) {
				return fmt.Errorf("learning_progress.last_opened_at 无效")
			}
		}
	}
	return nil
}
