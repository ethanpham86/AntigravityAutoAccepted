# Architecture

## Directory Structure
- `cmd/main.go`: Entry point, loads config, initializes interactive selector and starts engine.
- `internal/capture`: Pure Windows API (`gdi32.dll`, `user32.dll`) screen region capture. Converts GDI bitmap into PNG temp files.
- `internal/clicker`: Pure Windows API (`user32.dll`) `SendInput` wrapper for moving the mouse and clicking.
- `internal/selector`: Simple Console-based coordinate collector.
- `internal/ocr`: Wrap `tesseract.exe` via CLI using TSV format to get accurate word bounding boxes.
- `internal/engine`: The infinite loop coordinator. Manages interval timing and logging statistics.

## Design Decisions
- **Zero CGO**: Avoiding CGO on Windows drastically simplifies compilation for typical users. Thus tools like `robotgo` or `gosseract` were skipped in favor of pure `syscall` wrappers and `tesseract` CLI wrappers.
- **Tesseract TSV Method**: `tesseract <img.png> stdout tsv` returns a tab-separated list of word boundaries. This allows translation of relative pixel coordinates (from a cropped region) back into absolute global Windows coordinate clicks.
