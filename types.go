package main

type Option struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Recommended bool   `json:"recommended,omitempty"`
}

type Question struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Options     []Option `json:"options"`
	MultiSelect bool     `json:"multiSelect,omitempty"`
	AllowCustom *bool    `json:"allowCustom,omitempty"`
}

type AskQuestionsInput struct {
	Questions []Question `json:"questions" jsonschema:"Questions to present to the user"`
}

type Session struct {
	ID        string     `json:"id"`
	Questions []Question `json:"questions"`
}

type QuestionAnswer struct {
	SelectedIndexes []int  `json:"selectedIndexes"`
	CustomText      string `json:"customText,omitempty"`
}

type SessionAnswer struct {
	Answers []QuestionAnswer `json:"answers"`
}
