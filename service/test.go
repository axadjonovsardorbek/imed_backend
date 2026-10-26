package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	m "imed/pkg/models"
	st "imed/storage/postgres"

	"github.com/xuri/excelize/v2"
)

type TestService struct {
	storage st.Storage
}

func NewTestService(storage *st.Storage) *TestService {
	return &TestService{storage: *storage}
}

func (s *TestService) Create(ctx context.Context, req *m.TestCreateReq) (*m.ById, error) {
	return s.storage.TestS.Create(ctx, req)
}
func (s *TestService) GetById(ctx context.Context, req *m.ById) (*m.TestRes, error) {
	return s.storage.TestS.GetById(ctx, req)
}
func (s *TestService) GetAll(ctx context.Context, req *m.TestGetListReq) (*m.PaginationResponse, error) {
	return s.storage.TestS.GetAll(ctx, req)
}
func (s *TestService) Update(ctx context.Context, req *m.TestUpdateReq) (*m.Void, error) {
	return s.storage.TestS.Update(ctx, req)
}
func (s *TestService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.TestS.Delete(ctx, req)
}
func (s *TestService) AddQuestion(ctx context.Context, req *m.TestQuestionReq) (*m.ById, error) {
	ids, err := s.storage.TestS.AddQuestions(ctx, req.TestId, []m.TestQuestionBody{req.TestQuestionBody})
	if err != nil {
		return nil, err
	}
	return &m.ById{Id: ids[0]}, nil
}
func (s *TestService) UpdateQuestion(ctx context.Context, req *m.TestQuestionReq) (*m.Void, error) {
	return s.storage.TestS.UpdateQuestion(ctx, req)
}
func (s *TestService) DeleteQuestion(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.TestS.DeleteQuestion(ctx, req)
}

// Import Excel fayldagi savollarni testga qo'shadi.
// Format: 1-qator sarlavha; keyingi qatorlar: savol | A | B | C | ... | to'g'ri javob (harf).
// To'g'ri javob ustuni — sarlavhadagi oxirgi ustun; variantlar soni sarlavha ustunlari bilan cheklanadi
// (ko'proq variant kerak bo'lsa, sarlavhaga ustun qo'shiladi).
func (s *TestService) Import(ctx context.Context, testId string, file io.Reader) (*m.TestImportRes, error) {
	questions, err := ParseTestExcel(file)
	if err != nil {
		return nil, err
	}
	ids, err := s.storage.TestS.AddQuestions(ctx, testId, questions)
	if err != nil {
		return nil, err
	}
	return &m.TestImportRes{Imported: len(ids)}, nil
}

func ParseTestExcel(file io.Reader) ([]m.TestQuestionBody, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("invalid excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("error while reading excel: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("excel file is empty")
	}
	header := trimRow(rows[0])
	answerCol := len(header) - 1
	if answerCol < 3 {
		return nil, fmt.Errorf("header must be: question | A | B | ... | correct answer")
	}

	var (
		questions []m.TestQuestionBody
		errs      []string
	)
	for i, row := range rows[1:] {
		line := i + 2
		cells := trimRow(row)
		if len(cells) == 0 {
			continue
		}
		if len(cells) > answerCol+1 {
			errs = append(errs, fmt.Sprintf("row %d: more columns than header", line))
			continue
		}
		for len(cells) <= answerCol {
			cells = append(cells, "")
		}

		optionCells := trimRow(cells[1:answerCol])
		answer := strings.ToUpper(cells[answerCol])
		switch {
		case cells[0] == "":
			errs = append(errs, fmt.Sprintf("row %d: question is empty", line))
			continue
		case len(optionCells) < 2:
			errs = append(errs, fmt.Sprintf("row %d: at least 2 options required", line))
			continue
		case len(answer) != 1 || answer[0] < 'A' || int(answer[0]-'A') >= len(optionCells):
			errs = append(errs, fmt.Sprintf("row %d: invalid correct answer %q", line, cells[answerCol]))
			continue
		}
		correctIdx := int(answer[0] - 'A')

		q := m.TestQuestionBody{Text: cells[0]}
		gap := false
		for j, text := range optionCells {
			if text == "" {
				gap = true
				break
			}
			q.Options = append(q.Options, m.TestOptionBody{Text: text, IsCorrect: j == correctIdx})
		}
		if gap {
			errs = append(errs, fmt.Sprintf("row %d: options must not have gaps", line))
			continue
		}
		questions = append(questions, q)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("excel file has no questions")
	}
	return questions, nil
}

// trimRow kataklarni tozalaydi va oxiridagi bo'sh kataklarni olib tashlaydi.
func trimRow(row []string) []string {
	cells := make([]string, 0, len(row))
	for _, c := range row {
		cells = append(cells, strings.TrimSpace(c))
	}
	for len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// ImportTemplate admin uchun namuna Excel fayl.
func (s *TestService) ImportTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	rows := [][]interface{}{
		{"Savol", "A", "B", "C", "D", "To'g'ri javob"},
		{"Yurak nechta kameradan iborat?", "2", "3", "4", "5", "C"},
		{"Eng katta organ qaysi?", "Jigar", "Teri", "O'pka", "", "B"},
	}
	for i, r := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &r); err != nil {
			return nil, err
		}
	}
	return f, nil
}

// ===== Mobile =====

func (s *TestService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.TestMobileRes, error) {
	return s.storage.TestS.GetByIdMobile(ctx, req)
}
func (s *TestService) Start(ctx context.Context, req *m.ByIdMobile) (*m.TestStartRes, error) {
	return s.storage.TestS.Start(ctx, req)
}
func (s *TestService) Answer(ctx context.Context, req *m.TestAnswerReq) (*m.TestAnswerRes, error) {
	return s.storage.TestS.Answer(ctx, req)
}
func (s *TestService) Finish(ctx context.Context, req *m.TestAttemptReq) (*m.TestFinishRes, error) {
	return s.storage.TestS.Finish(ctx, req)
}
func (s *TestService) History(ctx context.Context, req *m.TestHistoryReq) (*m.PaginationResponse, error) {
	return s.storage.TestS.History(ctx, req)
}
