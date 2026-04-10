# Go Expert Contract (Strict Mode)

## 1. ELK TRACKING [MANDATORY]
- **Continuous Logging:** Execute ELK Pre-Flight `_search`, Mid-Flight tracking (BEFORE & AFTER), and Post-Flight `_doc` POSTs natively.
- **Strict Schema:** Supply standard ELK fields (`question`, `question_keywords`, `solution`, `errors_encountered`, `error_fixes`, `files_modified`). Use "None" instead of null/blank. Avoid `[]`, use `["None"]`.

## 2. ARCHITECTURE & PERFORMANCE
- **Clean Architecture:** Enforce pure domain separation (`domain/`, `core/`, `modules/`, `pkg/`).
- **Database Pooling:** MUST implement `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`.
- **Constraint Rules:** No demo code. Strict error resolution. Clean up via `defer`. No global mutable states.

## 3. SECURITY, TESTING & QUALITY ASSURANCE
- **Code Review & Tuning:** Always self-review code for memory leaks, network bottlenecks, and data races implicitly.
- **InfoSec & Pentesting:** Architect securely. Eradicate SQL injection (use parameterized queries), XSS, and hardcoded secrets. Encrypt sensitive payloads unconditionally. Treat every endpoint as inherently hostile (Zero Trust).
- **Comprehensive Testing:** Mandatory Unit tests (`*_test.go`), integration tests, and simulated pentest boundary checks where applicable.
- **UI/UX Consistency:** If a Graphical Interface exists, strictly enforce UI/UX uniformity across the entire project (standardized modern aesthetics, global color palettes, micro-interactions, responsive frameworks).

## 4. DOCS & EXECUTION
- `README.md`, `ARCHITECTURE.md`, and `Makefile` are required.
- Auto-proceed: Implement robust code directly without iterative pauses to ask user permission.

## 5. GIT EXPORT
```bash
echo "code_*.tar.gz" > .gitignore && echo "*.sql" >> .gitignore
git init && git add . && git commit -m "feat/fix: msg"
git branch -M main && git remote add origin https://github.com/ethanpham86/CRM_N_BDS.git
git push -u origin main --force
```
