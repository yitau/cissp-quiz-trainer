package domain

const AppVersion = "0.1.2"
const DatabaseVersion = 2

type StoredQuestion struct {
	SetID    string
	Position int
	Question Question
}
type BackupContents struct {
	Documents []string
	Questions []StoredQuestion
	Sessions  []Session
	Settings  map[string]string
}
type BackupMetadata struct {
	FormatVersion   int    `json:"formatVersion"`
	AppVersion      string `json:"appVersion"`
	DatabaseVersion int    `json:"databaseVersion"`
	CreatedAt       string `json:"createdAt"`
	DatabaseSHA256  string `json:"databaseSHA256"`
	ConfigSHA256    string `json:"configSHA256"`
}
