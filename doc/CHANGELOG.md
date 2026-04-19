# Changelog

All notable changes to this project will be documented in this file.

## [v3.2.0] - 2026-04-19
### Added
- **Global Hotkey (F6/F7):** Nhấn F6 để Pause/Resume, F7 để Stop — hoạt động toàn hệ thống, không cần focus console.
- **`[CLICK]` Log Tag:** Mỗi click ghi rõ: keyword, toạ độ, phương thức, confidence. Format: `[CLICK] ✓ "ACCEPTED" @ (450,320) | OCR conf=95% | BG`.
- **Per-keyword Statistics:** Thống kê số click theo từng keyword khi thoát: `ACCEPTED: 12 clicks, OK: 3 clicks`.
- **`ocr.IsAvailable()`:** Check Tesseract 1 lần khi khởi động. Nếu không có → chạy template-matching-only, không spam error log.
- **DPI Awareness:** Gọi `SetProcessDPIAware()` khi khởi động → toạ độ chính xác trên high-DPI.
- **Portable Path Resolution:** Config, img, log resolve relative to exe location (không phải CWD). Copy folder sang máy khác là chạy.
- **New module `internal/hotkey/`:** Win32 `RegisterHotKey` + `GetMessage` pump.
- **`logger.Click()`:** Log function riêng cho click events, luôn hiện bất kể log level.

### Changed
- Engine click log format ngắn gọn hơn: `[CLICK] ✓ "keyword" @ (x,y) | method | BG`.
- Template match path giờ cũng log `Background/Physical` flag (trước đây chỉ OCR path log).
- Tesseract giờ là **optional** (trước đây bắt buộc, bị spam error nếu chưa cài).
- Version bump: v3.1.3 → v3.2.0.

---

## [v3.1.3] - 2026-04-16
### Added
- **Hybrid Dual-Path Architecture:** Template Matching (pixel SAD) + OCR Fallback.
- **Two-Pass Coarse-to-Fine Matching:** 200 pixel sub-grid → early exit, O(1) scaling.
- **Multi-threaded SAD:** 4 goroutines chia vùng Y cho template scanning.
- **Alpha Mask Tolerance:** Bỏ qua pixel transparent (alpha < 128).
- **Heuristic Observability:** Log `💡 Almost Matched (84.9%, Needs: 85%)`.
- **Non-Maximum Suppression:** 150px radius → 1 best match per button.
- **Background Click (PostMessage):** Không cướp chuột vật lý.

### Fixed
- Console window hiện đúng kể cả khi build `-H windowsgui`.
- Logger output chính xác khi stdout invalid (GUI subsystem).

---

## [v2.5] - 2026-04-10
### Added
- **Sweep Click:** Chuyển từ "Click 1 Hit" thành "Click All Matches". Click nhiều nút liền kề.
- **Coordinates Deduplication:** Bán kính 20px xung quanh Click.
- **File Physical Logging:** `autoclick.log` chuyên nghiệp (`internal/logger`).
- **Auto-Learner Guard:** Cảnh báo khi user dùng Full Screenshot thay vì cropped button.

### Fixed
- Toạ độ UIPI / Scaling — hệ quy chiếu điểm mù tuyệt đối.
- Chặn rác Console: Memory Leak warnings Leptonica via `debug_file=NUL`.
- Xoá `context.Canceled` warnings khi Graceful Shutdown.

---

## [v2.1] - 2026-04-09
### Changed
- Gỡ CGO. 100% Pure Go syscall (`user32` + `gdi32`).
- Keyword Matching: Chuyển sang Levenshtein Distance + Exact Substring Rule.

## [v1.0] - Pilot
### Added
- Selector Image.
- Loop Scan đơn lẻ.
- Initial codebase.
