package scoring

import (
	"encoding/csv"
	"os"
	"testing"
)

// TestGolden читает эталон математика. Пока файла нет, тест пропускается.
// TODO: когда придёт golden_cases.csv, разобрать колонки и сравнить PD, P_gap, точки А/Б.
func TestGolden(t *testing.T) {
	f, err := os.Open("testdata/golden_cases.csv")
	if os.IsNotExist(err) {
		t.Skip("testdata/golden_cases.csv ещё нет")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("эталонных строк: %d", len(rows)-1)
}
