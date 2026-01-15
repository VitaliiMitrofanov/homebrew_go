package main

import (
	"fmt"
	"log"
	"time"
)

type Transaction struct {
		Date        time.Time
		Description string
		Amount      string
		Balance     string
}

func saveTransactions(userID int64, bankName string, transactions []Transaction) error {
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions to save")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO statement_data (user_id, bank_name, transaction_date, description, amount, balance)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, txn := range transactions {
		_, err := stmt.Exec(
			userID,
			bankName,
			txn.Date,
			txn.Description,
			txn.Amount,
			txn.Balance,
		)
		if err != nil {
			log.Printf("Error inserting transaction: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Saved %d transactions for user %d, bank %s", len(transactions), userID, bankName)
	return nil
}
