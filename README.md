# AutoClickAccepted

A Windows Go program that continuously scans a user-selected screen region for target text (e.g., "Allow Once", "ACCEPTED", "RUN", "Retry") using OCR and automatically clicks on it.

## Features
- **OCR-based Text Detection**: Uses Tesseract OCR to find text anywhere on screen.
- **Pure Go Windows API**: Uses raw GDI and User32 `syscall`s. No CGO setup required.
- **Interactive Selector**: Start the app and select any region of the screen to monitor.
- **Configurable**: Define your target keywords and scan intervals in `config.yaml`.

## Prerequisites
1. **Windows OS**
2. **Go 1.21+**
3. **Tesseract OCR (Required)**
   - Download for Windows: [tesseract-ocr-w64-setup-5.3.3.20231005.exe](https://github.com/UB-Mannheim/tesseract/wiki)
   - Ensure `tesseract.exe` is added to your system `PATH` (e.g., `C:\Program Files\Tesseract-OCR\`).

## Setup and Installation

1. Clone or clone this repository to your Go workspace.
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the program directly:
   ```bash
   go run ./cmd/main.go
   ```
   Or build it:
   ```bash
   make build
   ./bin/autoclick.exe
   ```

## Usage
1. Modify `config.yaml` to set your target keywords (case-insensitive).
2. Start the program.
3. Follow the console prompts to select the **Top-Left** and **Bottom-Right** corners of the screen region you want to monitor.
4. The engine will loop and click automatically when the text appears! Stop it using `Ctrl+C`.
