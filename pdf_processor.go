package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coregx/gxpdf"
)

func processPDF(userID int64, bankName string, pdfData []byte) error {
	// Create a temporary file for gxpdf
	tmpFile, err := os.CreateTemp("", "statement_*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write PDF data to temp file
	if _, err := tmpFile.Write(pdfData); err != nil {
		return fmt.Errorf("failed to write PDF data: %w", err)
	}
	tmpFile.Close()

	tabulaPath := os.Getenv("TABULA_JAR_PATH")
	if tabulaPath == "" {
		tabulaPath = "tools/tabula/tabula.jar"
	}
	fmt.Printf("Tabula path: %s", tabulaPath)

	cmd := exec.Command(
		"java", "-jar", tabulaPath,
		"-p", "all",
		"-f", "JSON",
		"-t",
		tmpFile.Name(),
	)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract PDF: %w", err)
	}

	/*

	// Open PDF using gxpdf
	doc, err := gxpdf.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Extract tables from PDF
	tables := doc.ExtractTables()

	var allTransactions []Transaction

	// Process each table found in the PDF
	for _, table := range tables {
		rows := table.Rows()
		fmt.Printf("Table: %d rows x %d cols\n", table.RowCount(), table.ColumnCount())
		transactions := parseTable(bankName, rows)
		allTransactions = append(allTransactions, transactions...)

	}

	// Save transactions to database
	if err := saveTransactions(userID, bankName, allTransactions); err != nil {
		return fmt.Errorf("failed to save transactions: %w", err)
	}

	*/

	return nil
}

/*

// parseTable parses a table row and extracts transaction data
func parseTable(bankName string, row [][]string) []Transaction {

	// Use existing bank-specific parsers
	return parseTransactionsByBank(bankName, row)
}

type Transaction struct {
	Date        time.Time
	Description string
	Amount      string
	Balance     string
}

func parseTransactionsByBank(bankName string, row [][]string) []Transaction {
	switch bankName {
	case "Sberbank":
		return parseSberbank(row)
	default:
		return []Transaction{}
	}
}

func parseSberbank(rows [][]string) []Transaction {
	var transactions []Transaction
	var transaction Transaction
	dateTimePattern := "02.01.2006 15:04"
	for _, row := range rows {
		fmt.Printf("%#v\n", row)
		// row[0] = "02.01.2006"
		// row[1] = "15:04"
		// Складываем строки и проверяем. Если совпало с шаблоном времени, то это начало транзакции
		// Если index > 0, то это не первая транзакция
		// Добавляем предыдущую транзакцию в массив и начинаем новую

		dateTime, err := time.Parse(dateTimePattern, row[0]+" "+row[1])

		if err == nil {

			if transaction != (Transaction{}) {
				transactions = append(transactions, transaction)
			}

			transaction = Transaction{
				Date:    dateTime,
				Amount:  row[4],
				Balance: row[5],
			}

		} else {
			// Если ошибка при парсинге даты, то это описание в текущей транзакции
			// Добавляем его
			transaction.Description += " " + row[3]
		}

	}
	return transactions
}
*/