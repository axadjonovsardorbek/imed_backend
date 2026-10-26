package service

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func excelBytes(t *testing.T, rows [][]interface{}) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, r := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &r); err != nil {
			t.Fatal(err)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

func TestParseTestExcel(t *testing.T) {
	buf := excelBytes(t, [][]interface{}{
		{"Savol", "A", "B", "C", "D", "Javob"},
		{"Q1", "a", "b", "c", "d", "c"},
		{"Q2", "x", "y", "", "", "A"},
		{},
	})

	qs, err := ParseTestExcel(buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 2 {
		t.Fatalf("want 2 questions, got %d", len(qs))
	}
	if len(qs[0].Options) != 4 || !qs[0].Options[2].IsCorrect || qs[0].Options[0].IsCorrect {
		t.Fatalf("bad q1: %+v", qs[0])
	}
	if len(qs[1].Options) != 2 || !qs[1].Options[0].IsCorrect {
		t.Fatalf("bad q2: %+v", qs[1])
	}
}

func TestParseTestExcelErrors(t *testing.T) {
	cases := map[string][][]interface{}{
		"answer out of range": {{"Q", "A", "B", "Ans"}, {"Q", "a", "b", "C"}},
		"gap in options":      {{"Q", "A", "B", "C", "Ans"}, {"Q", "a", "", "c", "A"}},
		"too few options":     {{"Q", "A", "B", "Ans"}, {"Q", "a", "", "A"}},
		"no questions":        {{"Q", "A", "B", "Ans"}},
		"bad header":          {{"Q", "Ans"}, {"Q", "a", "b", "A"}},
	}
	for name, rows := range cases {
		if _, err := ParseTestExcel(excelBytes(t, rows)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestImportTemplateParses(t *testing.T) {
	f, err := (&TestService{}).ImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	buf, _ := f.WriteToBuffer()
	if _, err := ParseTestExcel(buf); err != nil {
		t.Fatalf("template must be importable: %v", err)
	}
}
