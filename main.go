package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	bot     *tgbotapi.BotAPI
	db      *sql.DB
	appName string
	appURL  string
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	appName = os.Getenv("APP_NAME")
	appURL = os.Getenv("APP_URL")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	if appName == "" {
		log.Fatal("APP_NAME is not set")
	}
	if appURL == "" {
		log.Fatal("APP_URL is not set")
	}

	// Initialize Telegram bot
	var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize database connection
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Check if we should use webhook or polling mode
	useWebhook := os.Getenv("USE_WEBHOOK")
	if useWebhook == "true" {
		// Check and set webhook
		if err := setupWebhook(); err != nil {
			log.Fatal("Failed to setup webhook:", err)
		}
		// Start HTTP server for webhook mode
		startWebhookServer()
	} else {
		// Delete webhook if exists and use polling mode
		if err := deleteWebhookIfExists(); err != nil {
			log.Printf("Warning: Failed to delete webhook: %v", err)
		}
		// Start listening for updates via polling
		startBot()
	}
}

func initDB() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "telegram_bot_db"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	// Create table if it doesn't exist
	if err := createTable(); err != nil {
		return err
	}

	log.Println("Database connection established")
	return nil
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS statement_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		bank_name VARCHAR(50) NOT NULL,
		transaction_date TIMESTAMP,
		description TEXT,
		amount VARCHAR(50) NULL,
		balance VARCHAR(50) NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_bank_name (bank_name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := db.Exec(query)
	return err
}

func setupWebhook() error {
	webhookURL := fmt.Sprintf("%s/webhook/%s", appURL, bot.Token)

	// Check current webhook
	webhookInfo, err := bot.GetWebhookInfo()
	if err != nil {
		return err
	}

	// If webhook URL doesn't match or doesn't exist, set it
	if webhookInfo.URL != webhookURL {
		// Delete existing webhook first if it exists
		if webhookInfo.URL != "" {
			deleteWebhook := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false}
			_, err = bot.Request(deleteWebhook)
			if err != nil {
				log.Printf("Warning: Failed to delete existing webhook: %v", err)
			}
		}

		// Set new webhook
		wh, err := tgbotapi.NewWebhook(webhookURL)
		if err != nil {
			return err
		}

		_, err = bot.Request(wh)
		if err != nil {
			return err
		}

		log.Printf("Webhook set to: %s", webhookURL)
	} else {
		log.Printf("Webhook already set correctly: %s", webhookURL)
	}

	return nil
}

func deleteWebhookIfExists() error {
	webhookInfo, err := bot.GetWebhookInfo()
	if err != nil {
		return err
	}

	if webhookInfo.URL != "" {
		log.Printf("Deleting existing webhook: %s", webhookInfo.URL)
		deleteWebhook := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false}
		_, err = bot.Request(deleteWebhook)
		if err != nil {
			return err
		}
		log.Println("Webhook deleted successfully")
	}

	return nil
}

func startBot() {
	log.Println("Starting bot in polling mode...")

	// Use polling mode
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallbackQuery(update.CallbackQuery)
		}
		if update.Message != nil {
			handleMessage(update.Message)
		}
	}
}

func startWebhookServer() {
	log.Println("Starting webhook server...")
	log.Println("Webhook mode requires an HTTP server. Please implement HTTP handler at:", fmt.Sprintf("%s/webhook/%s", appURL, bot.Token))
	log.Fatal("Webhook HTTP server not implemented. Set USE_WEBHOOK=false to use polling mode.")
}
