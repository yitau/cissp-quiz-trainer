package domain

type QuestionSet struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      int    `json:"domain,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}
type Question struct {
	ID             string            `json:"id"`
	Question       string            `json:"question"`
	Options        map[string]string `json:"options"`
	Answer         string            `json:"answer,omitempty"`
	Explanation    string            `json:"explanation,omitempty"`
	WhyCorrect     string            `json:"whyCorrect,omitempty"`
	WhyOthersWrong map[string]string `json:"whyOthersWrong,omitempty"`
	Domain         int               `json:"domain"`
	Type           string            `json:"type"`
	Difficulty     string            `json:"difficulty"`
	Tags           []string          `json:"tags,omitempty"`
}
type QuestionFile struct {
	SchemaVersion string      `json:"schemaVersion"`
	Set           QuestionSet `json:"set"`
	Questions     []Question  `json:"questions"`
}
type SetSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}
type Preview struct {
	Set       QuestionSet `json:"set"`
	Questions []Question  `json:"questions"`
	Status    string      `json:"status"`
	Messages  []string    `json:"messages"`
	Token     string      `json:"token"`
}
type Session struct {
	Statistics      Statistics    `json:"statistics"`
	DurationSeconds int64         `json:"durationSeconds"`
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Mode            string        `json:"mode"`
	Status          string        `json:"status"`
	StartedAt       string        `json:"startedAt"`
	CompletedAt     string        `json:"completedAt"`
	Items           []SessionItem `json:"items"`
}
type SessionItem struct {
	Question Question `json:"question"`
	Selected string   `json:"selected"`
	Flagged  bool     `json:"flagged"`
	Scored   bool     `json:"scored"`
	Correct  *bool    `json:"correct,omitempty"`
	ScoredAt string   `json:"scoredAt"`
}
type SessionSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	Total       int    `json:"total"`
	Scored      int    `json:"scored"`
	Correct     int    `json:"correct"`
}

// HideSolution is used at every answer-bearing application boundary.
func HideSolution(q Question) Question {
	q.Answer = ""
	q.Explanation = ""
	q.WhyCorrect = ""
	q.WhyOthersWrong = nil
	return q
}
