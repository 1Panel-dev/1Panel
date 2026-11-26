# Hướng dẫn cho AI agent (worker)

Repo này là **bản fork của 1Panel** (GPL-3.0), đóng gói lại với thương hiệu riêng (tên tạm **XPanel**) và tiếng Việt. Kế hoạch đã chốt; bạn là **worker**, làm đúng đặc tả từng task.

## Mỗi phiên làm việc
1. Đọc `fork-docs/PLAN.md` (bối cảnh, nguyên tắc), rồi `fork-docs/tasks/README.md` (quy trình).
2. Mở `fork-docs/tasks/STATUS.md`, chọn task theo quy tắc ở đầu file, đọc đặc tả trong `fork-docs/tasks/phase-*.md`.

## 6 quy tắc không được vi phạm
1. **Không sửa chuỗi "1Panel" trong mã nguồn upstream.** Đổi thương hiệu chỉ làm qua `scripts/brand/` lúc build.
2. **Không đổi tên nội bộ**: module Go, tên binary, `/opt/1panel`, `1pctl`, tên service.
3. Mọi sửa vào file của upstream phải là commit riêng, message bắt đầu bằng `[fork]`, và được liệt kê trong `fork-docs/PATCHES.md` (file nào, lý do, cách áp lại khi xung đột).
4. File mới của fork chỉ đặt trong: `branding/`, `scripts/brand/`, `scripts/release/`, `scripts/sync/`, `scripts/test/`, `.github/workflows/xp-*.yml`, `fork-docs/`. Riêng file dịch `vi` nằm cạnh các file ngôn ngữ khác.
5. Giữ nguyên `LICENSE`, mọi header bản quyền của FIT2CLOUD; không thêm code của bản Pro/XPack.
6. Không commit bí mật (token R2, khoá ký...). Dùng GitHub Secrets.

## Môi trường
- Repo ở `~/xpanel` **trong WSL2 Ubuntu 24.04** (không đặt trên ổ Windows `/mnt/c`, `/mnt/d`: chậm và dễ lỗi xuống dòng CRLF). Antigravity mở qua chế độ kết nối WSL.
- Đồng bộ và sao lưu bằng GitHub: **sau mỗi task phải `git push origin main`** (hoặc nhánh đang làm). Task chưa push coi như chưa xong.
- File `.sh` giữ xuống dòng LF.
- Go **1.27.3** (đúng như `core/go.mod`), Node **22 LTS** trở lên, npm, Docker (để test App Store), `make`, `upx` (tuỳ chọn).

## Kiểm tra chung trước khi báo xong
```bash
make build_all                      # frontend + core + agent
cd core && go vet ./... && go test ./... ; cd ../agent && go vet ./... && go test ./...
bash scripts/brand/apply.sh --check # từ T2.1 trở đi
node scripts/i18n/check-keys.mjs    # từ T3.1 trở đi
```
