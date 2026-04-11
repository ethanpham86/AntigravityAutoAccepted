# AutoClickAccepted v3.1.3 Hybrid (Fast & OCR Paths)

This project has been extensively upgraded into a **Hybrid Visual-OCR Engine** using pure Go dependencies and C++ level memory access. It captures arbitrary areas of your screen recursively.ements (e.g., exact template images from `img/` or target texts via OCR like "Allow Once", "ACCEPTED") and automatically clicks on them.

### Key Capabilities

*   **Hybrid Dual-Path Architecture:** Seamlessly fuses a **1:1 Native Pixel Template Matcher** with a **3x Upscaled Tesseract OCR** module. The engine falls back gracefully from visual to text recognition without user intervention.
*   **Nano-second Matching via RAM Pointer Bypass:** The `matcher` retrieves the native visual payload via `*image.RGBA` pixel matrices directly tied to Windows Desktop GDI bounds, bypassing `image.Image` reflection slowness. Runs at O(1) loop speeds.
*   **Two-Pass Coarse-to-Fine Grid:** Uses algorithmically identical sub-grids (200 pixels) to Early-Exit large template background noises resulting in a `1.5s` end-to-end processing loop per Region-of-Interest.
*   **Alpha Mask Tolerance:** Transparent components (`Alpha < 128`) are ignored in visual bounding checking.
*   **Heuristic Observability:** Notifies the console beautifully of `💡 Almost Matched (Similarity: 84.9%, Needs: 85%)` when your UI alters sub-pixels preventing silent fail mysteries.bbing your cursor!
- **OCR-based Text Detection Fallback**: Uses Tesseract OCR to find dynamic text anywhere on screen if image matching fails.
- **Pure Go Windows API**: Uses raw GDI and User32 `syscall`s. No CGO setup required.
- **Interactive Selector**: Start the app and select any region of the screen to monitor.
- **Configurable**: Define your target keywords and scan intervals in `config.yaml`.

## Prerequisites
1. **Windows OS**
2. **Go 1.21+**
3. **Tesseract OCR (v5.0+ installed and added to `%PATH%`)**
  > 💡 Tesseract OCR installer is bundled directly within this repository for offline setup. Run `dependencies/tesseract-ocr-w64-setup-5.5.0.20241111.exe` and complete the installation!

## Setup and Installation

1. Clone or download this repository to your Go workspace.
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run or Build the program:
   ```bash
   go build -o autoclick.exe ./cmd/main.go
   ```

## Usage
1. Modify `config.yaml` to set your target keywords (case-insensitive) and enable/disable `use_background_click`.
2. Start the `bin/autoclick_v3.1.exe` program.
3. Follow the console prompts to select the **Top-Left** and **Bottom-Right** corners of the screen region you want to monitor.
4. The engine will loop and click automatically when the text appears! Stop it using `Ctrl+C`.
