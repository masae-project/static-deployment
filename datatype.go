package main

type Solution struct {
	Name     string     `json:"name"`
	Projects *[]Project `json:"projects"`
	Run      *Run       `json:"run"`
	PreRun   *Run       `json:"prerun"`
}

type Project struct {
	Name     string         `json:"name"`
	Source   *string        `json:"source"`
	Run      *Run           `json:"run"`
	PreRun   *Run           `json:"prerun"`
	Replaces *[]ReplaceItem `json:"replaces"`
}

type ReplaceItem struct {
	Name   string          `json:"to"`
	Items  []ReplaceDetail `json:"items"`
	Run    *Run            `json:"run"`
	PreRun *Run            `json:"prerun"`
}

type Run struct {
	Default *[][]string `json:"default"`
	Windows *[][]string `json:"windows"`
	Linux   *[][]string `json:"linux"`
	Darwin  *[][]string `json:"darwin"`
}

type ReplaceDetail struct {
	Old string `json:"old"`
	New string `json:"new"`
	Num *int   `json:"num"`
}

type FileData struct {
	Path       string
	LoadString string
	LoadSize   int
	NewString  string
	NewSize    int
}

type Names struct {
	Solution string
	Project  string
	Replace  string
}

type BackupItem struct {
	SourceFile string
	JobName    Names
}
