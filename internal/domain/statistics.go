package domain

type Attempt struct {
	QuestionID string
	Domain     int
	Type       string
	Selected   string
	Correct    bool
	ScoredAt   string
}
type Metric struct {
	Label      string   `json:"label"`
	Count      int      `json:"count"`
	Correct    int      `json:"correct"`
	Wrong      int      `json:"wrong"`
	Unanswered int      `json:"unanswered"`
	Rate       *float64 `json:"rate"`
}
type Statistics struct {
	Total          Metric   `json:"total"`
	Domains        []Metric `json:"domains"`
	Types          []Metric `json:"types"`
	WrongQuestions int      `json:"wrongQuestions"`
	Favorites      int      `json:"favorites"`
}
type ReviewQuestion struct {
	Question       Question `json:"question"`
	Attempts       int      `json:"attempts"`
	Correct        int      `json:"correct"`
	Wrong          int      `json:"wrong"`
	Streak         int      `json:"streak"`
	Mastered       bool     `json:"mastered"`
	Favorite       bool     `json:"favorite"`
	LastAnsweredAt string   `json:"lastAnsweredAt"`
}
