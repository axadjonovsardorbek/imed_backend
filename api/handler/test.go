package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	m "imed/pkg/models"
	"imed/storage/postgres"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// testError foydalanuvchi xatolarini (vaqt tugagan, urinish yakunlangan) botga yubormaydi.
func testError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, postgres.ErrNoAccess):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case errors.Is(err, postgres.ErrAttemptNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, postgres.ErrAttemptFinished),
		errors.Is(err, postgres.ErrTestTimeUp),
		errors.Is(err, postgres.ErrInvalidOption):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Test error: ", err)
		smssender(c, err, http.StatusBadRequest)
	}
}

// CreateTest godoc
// @Summary Create a new test
// @Description Create a test with questions and options. Each question needs at least 2 options and exactly 1 correct option.
// @Tags test
// @Accept  json
// @Produce  json
// @Param test body models.TestCreateBody true "Test Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test [post]
func (h *Handler) CreateTest(c *gin.Context) {
	body := m.TestCreateBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if body.LessonId == "" {
		c.JSON(400, gin.H{"message": "lesson_id is required"})
		return
	}

	titleJSON, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "failed to serialize title"})
		return
	}

	res, err := h.Clients.Test.Create(context.Background(), &m.TestCreateReq{
		LessonId:     body.LessonId,
		Title:        string(titleJSON),
		TimeLimitSec: body.TimeLimitSec,
		OrderNum:     body.OrderNum,
		Questions:    body.Questions,
	})
	if err != nil {
		testError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "test created successfully",
		"Id":      res.Id,
	})
}

// GetByIdTest godoc
// @Summary Get test by ID
// @Description Get a test with questions and options (including correct answers)
// @Tags test
// @Produce  json
// @Param id path string true "Test ID"
// @Success 200 {object} models.TestRes
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/{id} [get]
func (h *Handler) GetByIdTest(c *gin.Context) {
	res, err := h.Clients.Test.GetById(context.Background(), &m.ById{Id: c.Param("id")})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// GetAllTests godoc
// @Summary Get all tests
// @Description Get tests, optionally filtered by lesson
// @Tags test
// @Produce  json
// @Param lesson_id query string false "Lesson ID"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.TestList
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test [get]
func (h *Handler) GetAllTests(c *gin.Context) {
	req := &m.TestGetListReq{
		LessonId: c.Query("lesson_id"),
		Filter:   m.ParsePagination(c.Query("limit"), c.Query("page")),
	}
	res, err := h.Clients.Test.GetAll(context.Background(), req)
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// UpdateTest godoc
// @Summary Update test
// @Description Update test title, time limit and order
// @Tags test
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Param test body models.TestUpdateBody true "Test update"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/{id}/update [put]
func (h *Handler) UpdateTest(c *gin.Context) {
	body := m.TestUpdateBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	titleJSON, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "failed to serialize title"})
		return
	}

	_, err = h.Clients.Test.Update(context.Background(), &m.TestUpdateReq{
		Id:           c.Param("id"),
		Title:        string(titleJSON),
		TimeLimitSec: body.TimeLimitSec,
		OrderNum:     body.OrderNum,
	})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "test updated successfully"})
}

// DeleteTest godoc
// @Summary Delete test
// @Tags test
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/{id}/delete [delete]
func (h *Handler) DeleteTest(c *gin.Context) {
	if _, err := h.Clients.Test.Delete(context.Background(), &m.ById{Id: c.Param("id")}); err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "test deleted successfully"})
}

// AddTestQuestion godoc
// @Summary Add question to test
// @Tags test
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Param question body models.TestQuestionBody true "Question"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/{id}/question [post]
func (h *Handler) AddTestQuestion(c *gin.Context) {
	body := m.TestQuestionBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	res, err := h.Clients.Test.AddQuestion(context.Background(), &m.TestQuestionReq{TestId: c.Param("id"), TestQuestionBody: body})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, gin.H{
		"message": "question created successfully",
		"Id":      res.Id,
	})
}

// UpdateTestQuestion godoc
// @Summary Update test question
// @Description Replaces question text and all options
// @Tags test
// @Accept json
// @Produce json
// @Param id path string true "Question ID"
// @Param question body models.TestQuestionBody true "Question"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/question/{id}/update [put]
func (h *Handler) UpdateTestQuestion(c *gin.Context) {
	body := m.TestQuestionBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if _, err := h.Clients.Test.UpdateQuestion(context.Background(), &m.TestQuestionReq{Id: c.Param("id"), TestQuestionBody: body}); err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "question updated successfully"})
}

// DeleteTestQuestion godoc
// @Summary Delete test question
// @Tags test
// @Produce json
// @Param id path string true "Question ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/question/{id}/delete [delete]
func (h *Handler) DeleteTestQuestion(c *gin.Context) {
	if _, err := h.Clients.Test.DeleteQuestion(context.Background(), &m.ById{Id: c.Param("id")}); err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "question deleted successfully"})
}

// ImportTestQuestions godoc
// @Summary Import questions from Excel
// @Description .xlsx: first row is header; columns: question | A | B | ... | correct letter. All rows are saved in one transaction.
// @Tags test
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Test ID"
// @Param file formData file true "Excel file"
// @Success 200 {object} models.TestImportRes
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/test/{id}/import [post]
func (h *Handler) ImportTestQuestions(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"message": "file is required"})
		return
	}
	file, err := fh.Open()
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	defer file.Close()

	res, err := h.Clients.Test.Import(context.Background(), c.Param("id"), file)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, res)
}

// TestImportTemplate godoc
// @Summary Download Excel import template
// @Tags test
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success 200 {file} file
// @Security BearerAuth
// @Router /web/test/import-template [get]
func (h *Handler) TestImportTemplate(c *gin.Context) {
	f, err := h.Clients.Test.ImportTemplate()
	if err != nil {
		testError(c, err)
		return
	}
	defer f.Close()

	c.Header("Content-Disposition", `attachment; filename="test_template.xlsx"`)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		slog.Error("Error writing template: ", err)
	}
}

// ===== Mobile =====

// GetByIdMobileTest godoc
// @Summary Get test for solving
// @Description Questions and options without correct answers
// @Tags test
// @Produce json
// @Param id path string true "Test ID"
// @Param Accept-Language header string false "uz | ru | en"
// @Success 200 {object} models.TestMobileRes
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /mobile/test/{id} [get]
func (h *Handler) GetByIdMobileTest(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	res, err := h.Clients.Test.GetByIdMobile(context.Background(), &m.ByIdMobile{Id: c.Param("id"), UserId: userId, Language: getLanguage(c)})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// StartTest godoc
// @Summary Start test attempt
// @Description Every call starts a new attempt ("Testni qayta ishlash" uses the same endpoint)
// @Tags test
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} models.TestStartRes
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /mobile/test/{id}/start [post]
func (h *Handler) StartTest(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	res, err := h.Clients.Test.Start(context.Background(), &m.ByIdMobile{Id: c.Param("id"), UserId: userId})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// AnswerTest godoc
// @Summary Answer a question
// @Description Returns immediately whether the answer is correct and the correct option. One answer per question.
// @Tags test
// @Accept json
// @Produce json
// @Param id path string true "Attempt ID"
// @Param answer body models.TestAnswerBody true "Answer"
// @Success 200 {object} models.TestAnswerRes
// @Failure 404 {object} string
// @Failure 409 {object} string "attempt finished / time is up / invalid option"
// @Security BearerAuth
// @Router /mobile/test/attempt/{id}/answer [post]
func (h *Handler) AnswerTest(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	body := m.TestAnswerBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	res, err := h.Clients.Test.Answer(context.Background(), &m.TestAnswerReq{
		AttemptId:  c.Param("id"),
		UserId:     userId,
		QuestionId: body.QuestionId,
		OptionId:   body.OptionId,
	})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// FinishTest godoc
// @Summary Finish test attempt
// @Description Returns result in percent and next lesson id. Unanswered questions count as wrong.
// @Tags test
// @Produce json
// @Param id path string true "Attempt ID"
// @Success 200 {object} models.TestFinishRes
// @Failure 404 {object} string
// @Security BearerAuth
// @Router /mobile/test/attempt/{id}/finish [post]
func (h *Handler) FinishTest(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	res, err := h.Clients.Test.Finish(context.Background(), &m.TestAttemptReq{AttemptId: c.Param("id"), UserId: userId})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}

// GetTestHistory godoc
// @Summary User test results history, grouped by test (best percent, attempts) with summary
// @Tags test
// @Produce json
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Param Accept-Language header string false "uz | ru | en"
// @Success 200 {object} models.TestHistoryRes
// @Security BearerAuth
// @Router /mobile/user/test-history [get]
func (h *Handler) GetTestHistory(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	res, err := h.Clients.Test.History(context.Background(), &m.TestHistoryReq{
		UserId:   userId,
		Language: getLanguage(c),
		Filter:   m.ParsePagination(c.Query("limit"), c.Query("page")),
	})
	if err != nil {
		testError(c, err)
		return
	}
	c.JSON(200, res)
}
