# Telegram Bank Statement Bot

A Telegram bot built in Go that processes bank statements from PDF files and saves transaction data to a MySQL database.

## Features

- `/statement` command to start processing bank statements
- Support for three banks: Sberbank, Yap, and Tbank
- PDF file processing and transaction extraction
- MySQL database storage
- Automatic webhook setup on startup

## Prerequisites

- Go 1.21 or higher
- MySQL database
- Telegram Bot Token (obtained from [@BotFather](https://t.me/botfather))

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd homebrew_go
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file based on `.env.example`:
```bash
cp .env.example .env
```

4. Edit `.env` and fill in your configuration:
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
APP_NAME=homebrew_go
APP_URL=https://your-domain.com
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=telegram_bot_db
```

5. Create the MySQL database:
```sql
CREATE DATABASE telegram_bot_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## Usage

### Running the Bot

```bash
go run .
```

Or build and run:
```bash
go build -o telegram-bot
./telegram-bot
```

### Bot Commands

- `/statement` - Start processing a bank statement

### Workflow

1. User sends `/statement` command
2. Bot presents three bank options: Sberbank, Yap, Tbank
3. User selects a bank (via inline keyboard or text)
4. Bot waits for PDF file
5. Bot processes PDF and extracts transaction data
6. Transaction data is saved to MySQL database

## Database Schema

The bot creates a `statement_data` table with the following structure:

```sql
CREATE TABLE statement_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    bank_name VARCHAR(50) NOT NULL,
    transaction_date DATE,
    description TEXT,
    amount DECIMAL(15,2),
    balance DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_bank_name (bank_name)
);
```

## PDF Processing

The bot extracts transaction data from PDF files using text parsing. Each bank has its own parser that recognizes different date formats and transaction structures:

- **Sberbank**: Date format `DD.MM.YYYY`
- **Yap**: Date format `DD/MM/YYYY` or `DD-MM-YYYY`
- **Tbank**: Date format `YYYY-MM-DD`

Note: The PDF parsers are simplified and may need adjustment based on the actual format of your bank statements.

## Webhook Setup

On startup, the bot automatically:
1. Checks if a webhook is already set
2. Compares the webhook URL with `APP_URL/webhook/{bot_token}`
3. Sets or updates the webhook if needed

For production, you should set up an HTTP server to receive webhook updates. The current implementation uses polling mode for development.

## Project Structure

```
homebrew_go/
├── main.go           # Application entry point, webhook setup, database initialization
├── handlers.go       # Message and callback handlers
├── pdf_processor.go  # PDF parsing and transaction extraction
├── database.go       # Database operations
├── go.mod            # Go module dependencies
├── .env.example      # Environment variables template
└── README.md         # This file
```

## Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/joho/godotenv` - Environment variable management
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/gen2brain/go-fitz` - PDF text extraction (MuPDF wrapper)

## License

MIT
