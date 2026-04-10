---
name: elasticsearch-knowledge-base
description: Absolute strict 6-Layer ELK AI workflow.
---
# 🧠 ELK AI Memory Skill
**Host:** `https://127.0.0.1:9200` | **Creds:** `elastic:123456` | **Flag:** `-k`

> [!IMPORTANT]
> **AI ZERO-TOLERANCE OVERRIDE:**
> Khi khởi tạo phiên mới và được yêu cầu Nạp Skill này, bạn PHẢI tự rà soát và bắt đầu Thought Process bằng câu: *"I MUST COMPLY WITH ELK LOGGING RULES"*.
> Bạn **TUYỆT ĐỐI BỊ CẤM** gọi bất cứ Tool thao tác nào (`run_command` thay đổi code, `write_to_file`, `replace_file_content`) nếu CHƯA hoàn thành lệnh ELK `_search` và `_doc` POST khởi tạo (Pre-flight). Vi phạm điều này là vi phạm nghiêm trọng hợp đồng.

## 1. ABSOLUTE RULES
1. **PRE-FLIGHT SEARCH:** Call `_search` BEFORE writing code/debugging.
2. **MID-FLIGHT LOGGING:** Log EVERY tool call (Before & After actions) via `POST`.
   - **CRITICAL SCHEMA:** Use ONLY the standard schema. NULL values strictly forbidden. If an array is empty, use `["None"]` instead of `[]`. Map actions into `question`, `solution`, `errors_encountered`, `error_fixes`, `files_modified`.
3. **POST-FLIGHT LOGGING:** At task end, POST full summary context.
4. **SECURITY:** Mask sensitive keys (no real passwords).

> [!CAUTION]
> **MANDATORY RICH DESCRIPTION RULE:**
> Các field `solution`, `root_cause_analysis`, `error_fixes` **PHẢI chứa mô tả kỹ thuật chi tiết** (3-5 câu trở lên), giải thích rõ:
> - **Nguyên nhân gốc (Root Cause):** Tại sao lỗi xảy ra?
> - **Phân tích kỹ thuật:** Dữ liệu/logic nào sai? (ví dụ: "Tesseract trả conf dạng float 96.33 nhưng code dùng Atoi chỉ parse int → 100% từ bị skip")
> - **Giải pháp chi tiết:** Sửa cụ thể gì, đổi API nào, thêm logic gì?
> - **Impact:** Fix này ảnh hưởng module nào, có side effect không?
>
> **TUYỆT ĐỐI CẤM** ghi mô tả ngắn 1 dòng dạng "Fixed the bug" hoặc "Changed X to Y". Vi phạm = hợp đồng bị phá vỡ.

## 2. STANDARD JSON SCHEMA (Mandatory)
```json
{
  "@timestamp": "UTC Time",
  "project": "Project",
  "conversation_id": "Context ID",
  "question": "What is the intent/problem?",
  "question_keywords": ["keys", "tags"],
  "category": "bug-fix|feature|migration|config",
  "tags": ["go", "elk"],
  "root_cause_analysis": "Chi tiết phân tích nguyên nhân gốc rễ: dữ liệu sai gì, logic nào hỏng, environment nào gây ra. Viết ≥3 câu.",
  "solution": "Hành động cụ thể đã thực hiện: thay đổi API nào, thêm logic gì, tại sao chọn cách này. Viết ≥3 câu.",
  "solution_code": "Code snippet minh hoạ hoặc N/A",
  "errors_encountered": "Nội dung lỗi/log gốc copy nguyên bản, hoặc None",
  "error_fixes": "Chiến lược sửa lỗi chi tiết: bước 1 → bước 2 → kết quả. Viết ≥2 câu.",
  "impact_scope": "Danh sách module/file bị ảnh hưởng và side effect tiềm ẩn",
  "files_modified": ["None"],
  "status": "pending|resolved",
  "created_at": "ISO Time",
  "updated_at": "ISO Time"
}
```

## 3. POWERSHELL CURL WORKFLOW
**Always write JSON buffer to `C:\tmp\xxx.json` before sending.**
- **GET:** `curl.exe -k -ns -u elastic:123456 -X GET "https://127.0.0.1:9200/ai_knowledge_base/_search" -H "Content-Type: application/json" -d @C:\tmp\elk_search.json`
- **POST:** `curl.exe -k -ns -u elastic:123456 -X POST "https://127.0.0.1:9200/ai_knowledge_base/_doc" -H "Content-Type: application/json" -d @C:\tmp\elk.json`
