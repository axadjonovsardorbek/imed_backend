package models

// ===== Admin =====

type TestOptionBody struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type TestQuestionBody struct {
	Text     string           `json:"text"`
	ImageUrl string           `json:"image_url"`
	OrderNum int              `json:"order_num"`
	Options  []TestOptionBody `json:"options"`
}

type TestCreateBody struct {
	LessonId     string             `json:"lesson_id"`
	Title        map[string]string  `json:"title"`
	TimeLimitSec *int               `json:"time_limit_sec"`
	OrderNum     int                `json:"order_num"`
	Questions    []TestQuestionBody `json:"questions"`
}

type TestCreateReq struct {
	LessonId     string
	Title        string
	TimeLimitSec *int
	OrderNum     int
	Questions    []TestQuestionBody
}

type TestUpdateBody struct {
	Title        map[string]string `json:"title"`
	TimeLimitSec *int              `json:"time_limit_sec"`
	OrderNum     int               `json:"order_num"`
}

type TestUpdateReq struct {
	Id           string
	Title        string
	TimeLimitSec *int
	OrderNum     int
}

type TestQuestionReq struct {
	Id     string // question id (update)
	TestId string // test id (create)
	TestQuestionBody
}

type TestOptionRes struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	OrderNum  int    `json:"order_num"`
}

type TestQuestionRes struct {
	Id       string          `json:"id"`
	Text     string          `json:"text"`
	ImageUrl *string         `json:"image_url"`
	OrderNum int             `json:"order_num"`
	Options  []TestOptionRes `json:"options"`
}

type TestRes struct {
	Id            string            `json:"id"`
	LessonId      string            `json:"lesson_id"`
	Title         map[string]string `json:"title"`
	TimeLimitSec  *int              `json:"time_limit_sec"`
	OrderNum      int               `json:"order_num"`
	QuestionCount int               `json:"question_count"`
	Questions     []TestQuestionRes `json:"questions,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type TestGetListReq struct {
	LessonId string
	Filter   Pagination
}

type TestList struct {
	Tests []TestRes `json:"tests"`
	Count int       `json:"count"`
}

type TestImportRes struct {
	Imported int `json:"imported"`
}

// ===== Mobile =====

type TestOptionMobile struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type TestQuestionMobile struct {
	Id       string             `json:"id"`
	Text     string             `json:"text"`
	ImageUrl *string            `json:"image_url"`
	Options  []TestOptionMobile `json:"options"`
}

type TestMobileRes struct {
	Id           string               `json:"id"`
	LessonId     string               `json:"lesson_id"`
	Title        string               `json:"title"`
	TimeLimitSec *int                 `json:"time_limit_sec"`
	Questions    []TestQuestionMobile `json:"questions"`
	BestPercent  *float64             `json:"best_percent"`
}

type TestShortMobile struct {
	Id            string   `json:"id"`
	Title         string   `json:"title"`
	TimeLimitSec  *int     `json:"time_limit_sec"`
	QuestionCount int      `json:"question_count"`
	BestPercent   *float64 `json:"best_percent"`
}

type TestStartRes struct {
	AttemptId    string `json:"attempt_id"`
	StartedAt    string `json:"started_at"`
	TimeLimitSec *int   `json:"time_limit_sec"`
}

type TestAnswerBody struct {
	QuestionId string `json:"question_id"`
	OptionId   string `json:"option_id"`
}

type TestAnswerReq struct {
	AttemptId  string
	UserId     string
	QuestionId string
	OptionId   string
}

type TestAnswerRes struct {
	IsCorrect       bool   `json:"is_correct"`
	CorrectOptionId string `json:"correct_option_id"`
}

type TestAttemptReq struct {
	AttemptId string
	UserId    string
}

type TestFinishRes struct {
	AttemptId    string  `json:"attempt_id"`
	TestId       string  `json:"test_id"`
	CorrectCount int     `json:"correct_count"`
	TotalCount   int     `json:"total_count"`
	Percent      float64 `json:"percent"`
	NextLessonId *string `json:"next_lesson_id"`
}

type TestHistoryReq struct {
	UserId   string
	Language string
	Filter   Pagination
}

type TestHistoryItem struct {
	AttemptId    string  `json:"attempt_id"`
	TestId       string  `json:"test_id"`
	TestTitle    string  `json:"test_title"`
	LessonId     string  `json:"lesson_id"`
	CorrectCount int     `json:"correct_count"`
	TotalCount   int     `json:"total_count"`
	Percent      float64 `json:"percent"`
	FinishedAt   string  `json:"finished_at"`
}
