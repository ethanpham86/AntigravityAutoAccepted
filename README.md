# AutoClickAccepted 🤖🎯

AutoClickAccepted is a high-performance, purely Go-native auto-clicker built for Windows. It utilizes the power of **Tesseract OCR** and direct **Windows GDI/User32 syscalls** to visually scan specific regions of the screen for predefined keywords (e.g., "ACCEPTED", "Allow", "Run") and clicks them instantly without hijacking the system mouse coordinates negatively.

It was engineered to solve issues with UIPI (User Interface Privilege Isolation), multi-monitor scaling, and aggressive OCR false-positives.

![GitHub release](https://img.shields.io/badge/version-v2.5-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg)

---

## 🔥 Key Features

1. **Sweep Click (Multi-Targeting)**
   - Scans the screen and instantly maps ALL matching targets.
   - Intelligent de-duplication (radius 20px) to prevent double-clicking the same UI element.
   - Clears out all popup dialogs consecutively in one sweep (with 500ms intervals between clicks).
   
2. **Priority-based Cognitive Engine**
   - Config-driven `Keywords` priority (Top-down order in `config.yaml`).
   - Instead of blindly clicking the "longest" word, the engine prioritizes exact operational intents.

3. **Auto-Learner & Memory**
   - Place small cropped images of buttons into the `img/` directory.
   - Engine dynamically extracts words from `img/` and adds them to its cognitive dictionary on boot.
   - *(Note: ONLY use tightly cropped images of the buttons. Do not put full-screen screenshots in `img/`)*.

4. **Advanced Logging & Observability**
   - Real-time event traces mirrored to console and `autoclick.log`.
   - Tunable via `log_level` (`info`, `debug`, `error`) in `config.yaml`.
   - Memory Leak Warnings from Tesseract C++ ObjectCache are completely suppressed (`debug_file=NUL`).

5. **Anti-UIPI Payload Strategy**
   - Injects virtual inputs using `user32.dll` directly to the active Window loop.
   - Works flawlessly across multi-monitor setups with negative coordinates.

---

## ⚡ Getting Started

### 1. Prerequisites 
- **Golang** (v1.22+)
- **Tesseract OCR** (v5.0+ installed and added to `%PATH%`)
  > Download from: [https://github.com/UB-Mannheim/tesseract/wiki](https://github.com/UB-Mannheim/tesseract/wiki)

### 2. Prepare Config & Folders
If you clone a fresh repository, you need to manually scaffold the missing skeleton elements:
- Create a file named **`config.yaml`** at the root of the project with your desired operational variables (Refer to `doc/USAGE.md`).
- Create an empty folder named **`img/`**.
  - **What are the files in `img/`?**: Whenever you drop cropped images (`.png`) of buttons into this folder, they act as "Visual Templates". At startup, the Bot's Auto-Learner engine scans these images, extracts the text inside them via OCR, and memorizes those words directly into its active hunting dictionary (effectively merging them without needing to type them in `config.yaml`).

### 3. Build & Run
First, compile the application into a standalone executable:
```bash
go build -o bin/autoclick.exe ./cmd/main.go
```
Run the compiled binary (or use `go run`):
```bash
./bin/autoclick.exe
```

### 4. Usage & Configurations
Read the detailed [USAGE.md](doc/USAGE.md) for instructions on defining keywords in `config.yaml` and selecting screen regions.

For the architectural design philosophy, reference [ARCHITECTURE.md](doc/ARCHITECTURE.md).

---

## 📞 Troubleshooting & Graceful Exit
- **Stop safely**: Press `Ctrl+C` at any time over the terminal. The engine handles the `STATUS_CONTROL_C_EXIT` (`0xc000013a`) signal gracefully without throwing false-positive errors.
- **Run as Administrator**: If the target application is running as admin, the terminal used to execute this bot MUST also be "Run as Administrator" due to Windows UIPI security rules.

## 📜 Documentation
- [`doc/ARCHITECTURE.md`](doc/ARCHITECTURE.md) - Deep dive into system modules.
- [`doc/USAGE.md`](doc/USAGE.md) - Configuration manual and tips.
- [`doc/CHANGELOG.md`](doc/CHANGELOG.md) - Version history and bug fixes.
