package domain

type LessonFile struct {
	Format        string          `json:"format"`
	SchemaVersion string          `json:"schemaVersion"`
	Lesson        Lesson          `json:"lesson"`
	Sections      []LessonSection `json:"sections"`
	Concepts      []Concept       `json:"concepts"`
	CaseStudy     *CaseStudy      `json:"caseStudy,omitempty"`
}
type Lesson struct {
	ID                  string  `json:"id"`
	Title               string  `json:"title"`
	Description         string  `json:"description,omitempty"`
	Week                int     `json:"week"`
	Day                 int     `json:"day"`
	Domain              int     `json:"domain"`
	CreatedAt           string  `json:"createdAt"`
	LinkedQuestionSetID *string `json:"linkedQuestionSetId,omitempty"`
}
type LessonSection struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	ConceptIDs []string `json:"conceptIds"`
}
type Term struct {
	EN string `json:"en"`
	ZH string `json:"zh"`
}
type Recall struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
}
type Concept struct {
	ID            string   `json:"id"`
	SectionID     string   `json:"sectionId"`
	Term          Term     `json:"term"`
	Objective     string   `json:"objective"`
	Definition    string   `json:"definition"`
	Explanation   string   `json:"explanation"`
	Analogy       string   `json:"analogy"`
	Scenario      string   `json:"scenario"`
	Controls      []string `json:"controls,omitempty"`
	CommonMistake string   `json:"commonMistake"`
	ExamTip       string   `json:"examTip"`
	Recall        Recall   `json:"recall"`
}
type CaseStep struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
}
type CaseStudy struct {
	Title      string     `json:"title"`
	Background string     `json:"background"`
	Steps      []CaseStep `json:"steps"`
	Takeaway   string     `json:"takeaway"`
}
type LessonProgress struct {
	UnitID       string `json:"unitId"`
	ConceptID    string `json:"conceptId"`
	Status       string `json:"status"`
	LastOpenedAt string `json:"lastOpenedAt"`
	UpdatedAt    string `json:"updatedAt"`
}
type LessonSummary struct {
	Lesson          Lesson `json:"lesson"`
	SectionCount    int    `json:"sectionCount"`
	ConceptCount    int    `json:"conceptCount"`
	Understood      int    `json:"understood"`
	Completed       bool   `json:"completed"`
	ResumeConceptID string `json:"resumeConceptId"`
}
type LessonView struct {
	Document        LessonFile       `json:"document"`
	Summary         LessonSummary    `json:"summary"`
	Progress        []LessonProgress `json:"progress"`
	LinkedSetStatus string           `json:"linkedSetStatus"`
}
type LessonPreview struct {
	Lesson       Lesson          `json:"lesson"`
	Sections     []LessonSection `json:"sections"`
	ConceptCount int             `json:"conceptCount"`
	Status       string          `json:"status"`
	Token        string          `json:"token"`
}
type StoredLesson struct {
	ID          string
	Document    string
	Fingerprint string
	ImportedAt  string
}
