package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserState represents the current state of a user in conversation
type UserState struct {
	WaitingForBank bool
	WaitingForPDF  bool
	SelectedBank   string
}

var userStates = make(map[int64]*UserState)

func handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID

	// Initialize user state if not exists
	if userStates[userID] == nil {
		userStates[userID] = &UserState{}
	}

	state := userStates[userID]

	// Handle /statement command
	if message.IsCommand() && message.Command() == "statement" {
		startStatementConversation(userID)
		return
	}

	// Handle bank selection
	if state.WaitingForBank {
		handleBankSelection(userID, message.Text)
		return
	}

	// Handle PDF file
	if state.WaitingForPDF && message.Document != nil {
		handlePDFFile(userID, message.Document)
		return
	}

	// Default response
	msg := tgbotapi.NewMessage(userID, "Please use /statement to start processing a bank statement.")
	bot.Send(msg)
}

func startStatementConversation(userID int64) {
	state := userStates[userID]
	state.WaitingForBank = true
	state.WaitingForPDF = false
	state.SelectedBank = ""

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Sberbank", "bank_sberbank"),
			tgbotapi.NewInlineKeyboardButtonData("Yap", "bank_yap"),
			tgbotapi.NewInlineKeyboardButtonData("Tbank", "bank_tbank"),
		),
	)

	msg := tgbotapi.NewMessage(userID, "Please select your bank:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data

	// Initialize user state if not exists
	if userStates[userID] == nil {
		userStates[userID] = &UserState{}
	}

	state := userStates[userID]

	// Handle bank selection from inline keyboard
	var bankName string
	switch data {
	case "bank_sberbank":
		bankName = "Sberbank"
	case "bank_yap":
		bankName = "Yap"
	case "bank_tbank":
		bankName = "Tbank"
	default:
		// Answer callback to remove loading state
		bot.Request(tgbotapi.NewCallback(callback.ID, "Invalid selection"))
		return
	}

	state.SelectedBank = bankName
	state.WaitingForBank = false
	state.WaitingForPDF = true

	// Answer callback query
	bot.Request(tgbotapi.NewCallback(callback.ID, fmt.Sprintf("Selected: %s", bankName)))

	// Edit message to show selection
	editMsg := tgbotapi.NewEditMessageText(userID, callback.Message.MessageID, fmt.Sprintf("You selected %s. Please send the PDF file.", bankName))
	bot.Send(editMsg)
}

func handleBankSelection(userID int64, bankText string) {
	state := userStates[userID]
	
	// Normalize bank name
	bankText = strings.TrimSpace(strings.ToLower(bankText))
	var bankName string
	
	switch bankText {
	case "sberbank", "1":
		bankName = "Sberbank"
	case "yap", "2":
		bankName = "Yap"
	case "tbank", "3":
		bankName = "Tbank"
	default:
		msg := tgbotapi.NewMessage(userID, "Invalid selection. Please choose: Sberbank, Yap, or Tbank")
		bot.Send(msg)
		return
	}

	state.SelectedBank = bankName
	state.WaitingForBank = false
	state.WaitingForPDF = true

	msg := tgbotapi.NewMessage(userID, fmt.Sprintf("You selected %s. Please send the PDF file.", bankName))
	bot.Send(msg)
}

func handlePDFFile(userID int64, document *tgbotapi.Document) {
	state := userStates[userID]

	if !state.WaitingForPDF {
		msg := tgbotapi.NewMessage(userID, "Please use /statement to start processing a bank statement.")
		bot.Send(msg)
		return
	}

	// Check if file is PDF
	if !strings.HasSuffix(strings.ToLower(document.FileName), ".pdf") {
		msg := tgbotapi.NewMessage(userID, "Please send a PDF file.")
		bot.Send(msg)
		return
	}

	// Download file
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: document.FileID})
	if err != nil {
		log.Printf("Error getting file: %v", err)
		msg := tgbotapi.NewMessage(userID, "Error downloading file. Please try again.")
		bot.Send(msg)
		return
	}

	// Get file content
	fileURL := file.Link(bot.Token)
	fileData, err := downloadFile(fileURL)
	if err != nil {
		log.Printf("Error downloading file content: %v", err)
		msg := tgbotapi.NewMessage(userID, "Error downloading file content. Please try again.")
		bot.Send(msg)
		return
	}

	// Process PDF based on selected bank
	msg := tgbotapi.NewMessage(userID, "Processing PDF file...")
	bot.Send(msg)

	if err := processPDF(userID, state.SelectedBank, fileData); err != nil {
		log.Printf("Error processing PDF: %v", err)
		msg := tgbotapi.NewMessage(userID, fmt.Sprintf("Error processing PDF: %v", err))
		bot.Send(msg)
		return
	}

	// Reset state
	state.WaitingForPDF = false
	state.WaitingForBank = false
	state.SelectedBank = ""

	msg = tgbotapi.NewMessage(userID, "PDF processed successfully! Data has been saved to the database.")
	bot.Send(msg)
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
