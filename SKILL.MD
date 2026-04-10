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
  "solution": "Action taken/Result",
  "solution_code": "Code chunk or N/A",
  "errors_encountered": "Errors/Logs or None",
  "error_fixes": "Fix strategy or N/A",
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
