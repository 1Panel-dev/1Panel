# Phase 0–2: Fork, build, phát hành, đổi thương hiệu

Mốc upstream: **1Panel v2.5.1**. Nếu khi bắt đầu upstream đã có tag mới hơn, vẫn làm trên v2.5.1 cho đến hết T1.5, sau đó đồng bộ theo T6.1.

---

## T0.1: Tạo fork và môi trường làm việc

**Phụ thuộc:** không. *(Bước 1 do chủ dự án làm trên GitHub.)*

**Việc cần làm**
1. *(Chủ dự án)* Fork 4 repo về tài khoản/tổ chức GitHub của mình, để **public**: `1Panel-dev/1Panel` → `xpanel`, `1Panel-dev/installer` → `xpanel-installer`, `1Panel-dev/docs` → `xpanel-docs`. (`appstore` để Phase 5.)
2. Trong WSL2: `git clone <url-fork> ~/xpanel`; `git remote add upstream https://github.com/1Panel-dev/1Panel.git`; `git fetch upstream --tags`. Kiểm tra `git config core.autocrlf` trả về rỗng hoặc `input` (không phải `true`).
   - Đặt danh tính commit: `git config user.name` / `user.email` theo chủ dự án; xác thực push bằng `gh auth login` (GitHub CLI) hoặc SSH key trong WSL.
3. Tạo nhánh `main` từ tag `v2.5.1` (`git switch -c main v2.5.1`), đặt làm nhánh mặc định trên GitHub. Nhánh `dev-v2` của upstream không dùng.
4. Copy thư mục kế hoạch vào repo: `fork-docs/PLAN.md`, `fork-docs/tasks/*`, `AGENTS.md` (gốc repo), và tạo `GEMINI.md` là bản sao `AGENTS.md`. Tạo `fork-docs/PATCHES.md` (bảng rỗng: `| Commit | File | Lý do | Cách áp lại |`).
5. Thêm vào cuối `README.md` của repo (đây là patch `[fork]` đầu tiên) một đoạn: "Bản phân phối XPanel dựa trên 1Panel của FIT2CLOUD, giấy phép GPL-3.0. Mã nguồn: <link repo>." Không xoá nội dung README gốc.
6. Tắt các workflow của upstream không dùng được cho fork (`sync2gitee.yml`, `issue-translator.yml`, `add-labels-for-pr.yml`) bằng cách tắt trong tab Actions của GitHub (không xoá file, để tránh xung đột khi đồng bộ).

**Nghiệm thu**
- [ ] `git log --oneline -3` trên `main` cho thấy tag v2.5.1 + commit fork-docs + commit `[fork] README`.
- [ ] `git remote -v` có `origin` (fork) và `upstream`.

---

## T0.2: Build được từ mã nguồn

**Phụ thuộc:** T0.1.

**Việc cần làm**
1. Cài Go 1.27.3 (tarball chính thức vào `/usr/local/go`), Node 22 LTS (nvm), `make`, `curl`.
2. Chạy `make build_all`. Lưu ý:
   - `build_frontend` chạy `npm install && npm run build:pro`, kết quả vào `core/cmd/server/web/assets`.
   - `build_agent_on_linux` gọi `prepare_vm_helper`, script này tải `1panel-kvm-{amd64,arm64}` từ `resource.fit2cloud.com`. Ghi lại: có tải được không, kích thước, SHA256, và thông tin giấy phép nếu tìm thấy (để dùng cho T6.3).
3. Build cả `GOARCH=arm64` (core + agent).
4. Viết `fork-docs/BUILD.md`: các bước trên, thời gian build, kích thước binary, lỗi gặp phải và cách xử lý.

**Nghiệm thu**
- [ ] Có `build/1panel-core` và `build/1panel-agent` cho amd64 và arm64.
- [ ] `./build/1panel-core version` (hoặc lệnh tương đương mà binary hỗ trợ; tìm trong `core/cmd/server/cmd/`) in đúng phiên bản.
- [ ] `cd core && go test ./...` và `cd agent && go test ./...`: ghi số test pass/fail vào BUILD.md (test fail sẵn của upstream thì ghi lại, không sửa).

---

## T0.3: Phân tích gói phát hành và luồng cài đặt của upstream

**Phụ thuộc:** T0.1.

**Việc cần làm**
1. Tải bản chính thức để so sánh: `curl -s https://resource.1panel.pro/v2/stable/latest` → tải `.../v2/stable/<ver>/release/1panel-<ver>-linux-amd64.tar.gz` và `checksums.txt`.
2. Giải nén, ghi lại **chính xác** cấu trúc thư mục: tên thư mục gốc, các file (`1panel-core`, `1panel-agent`, `1pctl`, `install.sh`, `initscript/*`, `lang/*`, file khác nếu có), quyền file.
3. Đọc `quick_start.sh` và `install.sh` của repo installer (nhánh `v2`): ghi lại mọi URL được gọi, mọi file được ghi ra hệ thống, file edition (`$EDITION_FILE`), và cách `1pctl` biết phiên bản.
4. Trong mã nguồn core/agent: tìm luồng "kiểm tra cập nhật" và "nâng cấp" (từ khoá `RepoURL()`, `upgrade`): URL nào được gọi, file nào được tải, kiểm tra checksum ra sao, các file nào bị thay khi nâng cấp.
5. Viết `fork-docs/RELEASE-FORMAT.md` gồm 3 phần trên, kèm sơ đồ thư mục trên máy chủ phát hành:
   ```
   /v2/stable/latest                                    (nội dung: v2.5.1)
   /v2/stable/<ver>/release/1panel-<ver>-linux-<arch>.tar.gz
   /v2/stable/<ver>/release/checksums.txt
   /v2/<...các đường dẫn khác mà upgrade/ResourceURL dùng...>
   ```

**Nghiệm thu**
- [ ] RELEASE-FORMAT.md đủ để một người khác tự dựng máy chủ phát hành mà không phải đọc code.

---

## T1.1: CI kiểm tra build

**Phụ thuộc:** T0.2.

**File:** `.github/workflows/xp-ci.yml`.

**Đặc tả:** chạy khi push/PR vào `main`: setup Go (đọc phiên bản từ `core/go.mod` bằng `go-version-file`), Node 22, cache npm + Go; `make build_all`; `go vet` + `go test` cho core và agent (test fail sẵn của upstream ghi trong BUILD.md thì loại trừ bằng `-skip`, liệt kê rõ trong workflow); từ T2.1 thêm bước `scripts/brand/apply.sh --check`; từ T3.4 thêm `node scripts/i18n/check-keys.mjs`.

**Nghiệm thu:** [ ] Workflow xanh trên `main`. [ ] Thời gian chạy < 20 phút.

---

## T1.2: Máy chủ phát hành và workflow release

**Phụ thuộc:** T0.3, T1.1. *(Chủ dự án tạo bucket R2 + tên miền `dl.<domain>`, thêm secrets `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`.)*

**File:** `.github/workflows/xp-release.yml`, `scripts/release/package.sh`, `scripts/release/upload.sh`.

**Đặc tả**
- Kích hoạt khi push tag `v*-xp.*`.
- `package.sh <version>`: với mỗi arch (`amd64`, `arm64`): build core + agent, lấy `1pctl`, `install.sh`, `initscript/`, `lang/` từ repo `xpanel-installer` (checkout cùng workflow, tag/nhánh ghi trong biến `INSTALLER_REF`), đóng gói **đúng cấu trúc** ghi trong RELEASE-FORMAT.md, tên `1panel-<version>-linux-<arch>.tar.gz`; sinh `checksums.txt` (sha256, cùng định dạng upstream).
- `upload.sh`: dùng `aws s3 cp --endpoint-url https://<account>.r2.cloudflarestorage.com` để tải lên `v2/stable/<version>/release/`, sau cùng mới ghi `v2/stable/latest` (để người đang cài không lấy phải bản chưa upload xong). Tag có `-rc` thì đẩy vào kênh `dev` thay vì `stable`.
- Đồng thời tạo GitHub Release kèm các file + ghi chú "Mã nguồn: tag này" (đáp ứng GPL).

**Nghiệm thu:** [ ] Push tag `v2.5.1-xp.0-rc1` → file xuất hiện đúng đường dẫn trên `dl.<domain>/v2/dev/...`. [ ] `checksums.txt` khớp.

---

## T1.3: URL máy chủ đọc từ cấu hình `[fork]`

**Phụ thuộc:** T0.3.

**File sửa:** `core/global/global.go`, `agent/global/global.go` (+ nơi định nghĩa struct config `CONF.Base`); ghi vào PATCHES.md.

**Đặc tả**
- Thêm 3 trường cấu hình tuỳ chọn trong `Base` (đặt tên theo đúng kiểu đặt tên của struct đó, ví dụ `RepoURL`, `ResourceURL`, `AppRepoURL`) đọc từ file cấu hình hiện có của panel (tìm nơi `CONF` được nạp) **và** từ biến môi trường `XP_REPO_URL`, `XP_RESOURCE_URL`, `XP_APP_REPO_URL`.
- Ba hàm `RepoURL()`, `ResourceURL()`, `AppRepoURL()`: nếu trường cấu hình khác rỗng → trả giá trị đó; ngược lại giữ nguyên logic upstream. **Chỉ thêm 3 dòng `if` ở đầu mỗi hàm**, không sửa logic cũ.
- Giá trị mặc định của bản XPanel được ghi vào cấu hình bởi installer (T1.4), không hard-code vào Go. Nhờ vậy bản build này vẫn chạy như upstream nếu không cấu hình.
- Unit test cho 3 hàm (có/không có override).

**Nghiệm thu:** [ ] Diff của commit `[fork]` chỉ gồm các dòng thêm (không có dòng `-` nào ngoài import nếu cần). [ ] Test pass.

---

## T1.4: Installer riêng

**Phụ thuộc:** T1.2, T1.3.

**Repo:** `xpanel-installer` (nhánh `v2` của fork; tạo nhánh `xp` từ đó và làm việc trên `xp`).

**Đặc tả**
- `quick_start.sh`: thay `https://resource.1panel.pro` bằng biến `XP_DL_BASE` (mặc định `https://dl.<domain>`) khai báo một lần ở đầu file.
- `install.sh`: sau bước ghi cấu hình, ghi 3 khoá URL ở T1.3 trỏ về `XP_DL_BASE` (RepoURL `…/v2`, ResourceURL `…/v2/resource`); `AppRepoURL` để trống (dùng App Store upstream) cho đến Phase 5.
- Edition: luôn ghi bản `intl` vào `$EDITION_FILE` (để logic `CONF.Base.Edition` của upstream không chọn máy chủ Trung Quốc khi chưa có override).
- Đổi chữ hiển thị cho người dùng (tên sản phẩm, link tài liệu, link diễn đàn) bằng biến ở đầu file `install.sh` (`PRODUCT_NAME`, `DOCS_URL`, `COMMUNITY_URL`); **giữ nguyên** tên binary, đường dẫn, tên service.
- Thêm `lang/vi.sh` (dịch từ `lang/en.sh`) và cho phép chọn ngôn ngữ `vi` (xem cách script chọn ngôn ngữ hiện tại).
- Mirror Docker: giữ danh sách upstream, thêm tuỳ chọn không dùng mirror nào khi máy ở ngoài Trung Quốc (kiểm tra logic hiện tại; nếu đã có thì chỉ ghi chú).
- Lệnh cài một dòng cuối cùng: `bash -c "$(curl -sSL https://dl.<domain>/v2/quick_start.sh)"`. `upload.sh` (T1.2) đẩy cả `quick_start.sh` lên `v2/quick_start.sh`.

**Nghiệm thu:** [ ] Cài thành công trên VM sạch: Ubuntu 22.04, Ubuntu 24.04, Debian 12, Rocky 9 (ghi bảng kết quả vào `fork-docs/TEST-MATRIX.md`). [ ] Sau khi cài, panel gọi `dl.<domain>` khi kiểm tra cập nhật (xem log hoặc `tcpdump`/log nginx của R2). [ ] Gỡ cài bằng `1pctl uninstall` sạch.

---

## T1.5: Nâng cấp qua nút trong panel

**Phụ thuộc:** T1.4.

**Việc cần làm**
1. Unit test (file mới `core/utils/common/version_xp_test.go`, không sửa file test upstream): `ComparePanelVersion` cho các cặp `v2.5.1-xp.2 > v2.5.1-xp.1`, `v2.5.2 > v2.5.1-xp.9`, `v2.5.1-xp.1 > v2.5.1`, `v2.6.0-xp.1 > v2.5.9-xp.3`. Hàm của agent cũng test tương tự. Nếu có cặp nào sai → QUESTIONS.md.
2. Phát hành `v2.5.1-xp.1` và `v2.5.1-xp.2` (chỉ khác số phiên bản). Cài bản `.1` trên VM, bấm nâng cấp trong panel → lên `.2`, dữ liệu (website, cài đặt) còn nguyên.
3. Kiểm tra trang "Giới thiệu / Phiên bản" hiển thị đúng; ghi chú bản phát hành (release notes) lấy từ đâu và hiện ra sao, cập nhật RELEASE-FORMAT.md.

**Nghiệm thu:** [ ] Nâng cấp qua UI thành công trên Ubuntu 22.04.

---

## T2.1: Kiểm kê và script đổi thương hiệu

**Phụ thuộc:** T0.2.

**File:** `branding/brand.env`, `branding/assets/` (logo, favicon, ảnh nền login), `branding/replacements.tsv`, `scripts/brand/apply.sh`, `scripts/brand/inventory.sh`, `fork-docs/BRANDING.md`.

**Đặc tả**
- `inventory.sh`: liệt kê mọi chỗ người dùng **nhìn thấy** chữ 1Panel/FIT2CLOUD hoặc link tới 1panel.pro/1panel.cn/fit2cloud.com: giá trị chuỗi trong `frontend/src/lang/modules/*.ts`, `core/i18n/lang/*.yaml`, `agent/i18n/lang/*.yaml`, `frontend/.env` (`VITE_GLOB_APP_TITLE`), `frontend/index.html`, ảnh trong `frontend/src/assets/images/` (`1panel-logo.svg`, `1panel-menu-logo.svg`, `1panel-login*.jpg/png`, `favicon*`), link cứng trong `.vue`/`.ts`. Xuất `fork-docs/BRANDING.md` dạng bảng: file · dòng · chuỗi · loại (UI text / link / ảnh / nội bộ – không đổi).
- `brand.env`: `BRAND_NAME=XPanel`, `BRAND_SITE=https://<domain>`, `BRAND_DOCS=https://docs.<domain>`, `BRAND_COMMUNITY=...`, `BRAND_GITHUB=https://github.com/<org>/xpanel`.
- `replacements.tsv`: cột `pattern` (chuỗi cố định, không regex) · `replacement` (có thể dùng `${BRAND_NAME}`...) · `glob` (file áp dụng). Ví dụ: `1Panel` → `${BRAND_NAME}` chỉ trong **giá trị** của file ngôn ngữ; `https://1panel.pro/docs` → `${BRAND_DOCS}`; …
- `apply.sh`:
  - Chỉ chạy trên bản sao làm việc trong CI hoặc khi có cờ `--yes` (không bao giờ commit kết quả; script từ chối chạy nếu thấy đang ở nhánh `main` mà không có `--yes`).
  - Bước 1: thay chuỗi theo `replacements.tsv`. Với file `.ts` ngôn ngữ và `.yaml`, chỉ thay trong **giá trị chuỗi**, không đụng tên key; dùng script Node (`scripts/brand/replace-i18n.mjs`) parse AST cho `.ts` (dùng `typescript` đã có trong devDependencies của frontend) và parser YAML có sẵn trong frontend `node_modules` (nếu không có thì dùng regex an toàn theo dòng `key: "value"`; ghi lựa chọn vào BRANDING.md).
  - Bước 2: chép đè ảnh từ `branding/assets/` (cùng tên file với upstream).
  - Bước 3: `--check` = chạy inventory lại, **fail** nếu còn chuỗi loại "UI text" hoặc "link" chưa đổi.
  - **Không** đổi: tên binary, `/opt/1panel`, `1pctl`, tên API, tên key i18n, tên module Go, URL GitHub dùng để tải mã nguồn upstream.
- Workflow release (T1.2) gọi `apply.sh --yes` trước khi build.

**Nghiệm thu**
- [ ] `git status` sạch sau khi chạy `inventory.sh`; `apply.sh` từ chối chạy trên main khi thiếu `--yes`.
- [ ] Build sau `apply.sh --yes`: trang login, tiêu đề tab, menu, trang Giới thiệu, email/thông báo hiện `XPanel`; ngôn ngữ zh/en/vi đều đổi.
- [ ] `apply.sh --check` pass trên bản đã đổi.

---

## T2.2: Ẩn quảng cáo bản Pro `[fork]`

**Phụ thuộc:** T2.1.

**Việc cần làm**
1. Với `PANEL_XPACK=false`, liệt kê nơi giao diện vẫn hiện mục/nút "Pro", "Upgrade to Pro", "企业版", link mua bản quyền (tìm trong `frontend/src/extensions/`, các component có `xpack`, `isProductPro`, `isMasterProductPro`, `/buy`, `license`).
2. Ẩn chúng bằng **một** patch nhỏ nhất có thể, ưu tiên: một hằng số/flag sẵn có → một chỗ kiểm tra tập trung → CSS. Ghi vào PATCHES.md.
3. Trang Giới thiệu: thêm dòng "Dựa trên 1Panel (FIT2CLOUD), GPL-3.0 · Mã nguồn" (link `BRAND_GITHUB`). Dòng này thêm qua brand script nếu được, nếu không thì là patch `[fork]`.

**Nghiệm thu:** [ ] Đi qua mọi mục menu: không còn lời mời mua Pro. [ ] Tính năng Community vẫn đủ.
