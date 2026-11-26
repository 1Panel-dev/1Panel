# Build XPanel (1Panel v2.5.1) từ mã nguồn

Ghi lại từ lần build ngày 2026-11-27 (T0.2), máy dev: WSL2 Ubuntu 26.04, 8 CPU, 15 GB RAM.

## 1. Công cụ

| Công cụ | Phiên bản đã dùng | Cài thế nào |
|---|---|---|
| Go | 1.27.3 (khớp `core/go.mod`, `agent/go.mod`) | Tarball chính thức `go1.27.3.linux-amd64.tar.gz` (SHA256 `65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`) giải nén vào `/usr/local/go`; thêm `/usr/local/go/bin` vào `PATH` trong `~/.bashrc` |
| Node | v22.23.5 (npm đi kèm) | `nvm install 22` |
| make | GNU Make 4.4.1 | `apt-get install make` |
| curl | có sẵn | dùng bởi `prepare_vm_helper` |

Không cần `gcc`: core và agent build với `CGO_ENABLED=0`.

## 2. Các bước

```bash
cd ~/xpanel
make build_all            # = build_frontend + build_core_on_linux + build_agent_on_linux (amd64)
git checkout frontend/package-lock.json   # npm install sửa file này, không được commit

# arm64 (dùng lại target của Makefile, ghi đè GOARCH và BUILD_PATH)
make GOARCH=arm64 BUILD_PATH=$HOME/xpanel/build/arm64 build_core_on_linux build_agent_on_linux
```

- `build_frontend`: `cd frontend && npm install && npm run build:pro`, kết quả vào `core/cmd/server/web/assets` (31 MB), được nhúng vào `1panel-core`.
- `build_agent_on_linux` gọi `prepare_vm_helper` (`agent/utils/vm/helper/prepare.sh`), tải KVM helper (xem mục 5) rồi nhúng vào `1panel-agent` bằng `go:embed`.
- Cờ build: `-trimpath -ldflags '-s -w'`, `CGO_ENABLED=0`, binary tĩnh, đã strip.

## 3. Thời gian và kích thước

| Bước | Thời gian (real) |
|---|---|
| `make build_all` lần đầu (npm install + vite + tải module Go) | 5 phút 36 giây (vite build riêng: 1 phút 6 giây) |
| core + agent arm64 (cache module đã có) | 2 phút 34 giây |
| `go test ./...` core / agent | 1 phút 34 giây / 1 phút 53 giây |

| File | Kích thước |
|---|---|
| `build/1panel-core` (amd64) | 81 662 112 B (~78 MB) |
| `build/1panel-agent` (amd64) | 84 234 400 B (~80 MB) |
| `build/arm64/1panel-core` | 78 971 040 B (~75 MB) |
| `build/arm64/1panel-agent` | 77 332 640 B (~74 MB) |

Chưa dùng `upx`.

## 4. Kiểm tra binary và phiên bản

- `go version -m build/1panel-core` → `go1.27.3`, module `github.com/1Panel-dev/1Panel/core (devel)`. `file` xác nhận ELF x86-64 / ARM aarch64, statically linked.
- **Binary không chứa số phiên bản.** `1panel-core version` (`core/cmd/server/cmd/version.go`) cần root, đọc `SystemVersion` trong `/opt/1panel/db/core.db`. Giá trị này được ghi lúc cài/khởi động từ biến `ORIGINAL_VERSION` trong `/usr/local/bin/1pctl` (`core/init/hook/hook.go`). `conf/app.yaml` nhúng trong binary chỉ có `version: v2.0.0`, `mode: dev`. ⇒ Phiên bản do installer/gói phát hành quyết định (T0.3, T1.2).
- Trên máy chưa cài, mọi lệnh của `1panel-core` (kể cả `--help`) **panic** ngay khi khởi tạo: `open /usr/local/bin/1pctl: no such file or directory` (`core/i18n/i18n.go:217`, gọi từ `init()` của `update.go`). Hành vi sẵn có của upstream, không sửa. Kiểm tra lệnh `version` thật phải làm trên máy đã cài (T1.4/T6.2).
- `1panel-agent --help` chạy được.

## 5. KVM helper (dữ liệu cho T6.3)

`prepare.sh` tải từ `https://resource.fit2cloud.com/1panel/resource/kvm/1panel-kvm-{amd64,arm64}` vào `agent/utils/vm/helper/bin/` (bị `.gitignore`, chỉ tải nếu file chưa có). Tải được, không cần xác thực.

| File | Kích thước | SHA256 |
|---|---|---|
| `1panel-kvm-amd64` | 4 438 784 B | `506a0724b65b50632198b2e2bf53d4063b8bbc0bfef15c1c2fca0536939b54b6` |
| `1panel-kvm-arm64` | 4 197 056 B | `f1008adbadbc7c10bb198a8dca7a40aafa9b053a486560606713fff3063fa9ab` |

- `Last-Modified` trên máy chủ: 2026-11-18. URL không có phiên bản ⇒ build lại sau này có thể ra helper khác; nên ghim SHA256 khi làm CI.
- Binary Go 1.27.3, **CGO_ENABLED=1, liên kết động** (glibc), tags `libvirt_dlopen,libvirt_without_qemu,libvirt_without_lxc`, module `github.com/1Panel-dev/1Panel/kvm` (gói `internal/runtime`, `internal/protocol`).
- Mã nguồn module `kvm` **không có** trong repo public 1Panel v2.5.1. Không tìm thấy chuỗi giấy phép/bản quyền nào trong binary.

## 6. Kết quả test và vet

`go vet ./...`: core và agent đều sạch.

| Module | Pass | Fail | Skip | Gói có test |
|---|---|---|---|---|
| core | 2 | 0 | 1 | 2 (75 gói không có test) |
| agent | 365 | 11 | 0 | 6 (115 gói không có test) |

Số đếm gồm cả subtest (`go test -json`). **Test fail sẵn của upstream** (không sửa), đều ở `agent/app/service`:

- `TestOfficialIconSameRevisionRefreshIsReadOnlyUntilSync` (`app_icon_sync_test.go:163`)
- `TestAppStoreIndexUsesCacheUntilRevisionChanges` và 9 subtest (`app_store_index_test.go:98/101`)

Nguyên nhân chung: test dựng server TLS giả (chứng chỉ `example.com`) nhưng code vẫn gọi thật `https://apps-assets.fit2cloud.com//1panel.json.version.txt` → `x509: certificate is valid for example.com ... not apps-assets.fit2cloud.com`. Trong CI (T1.1) loại trừ bằng:

```
-skip 'TestOfficialIconSameRevisionRefreshIsReadOnlyUntilSync|TestAppStoreIndexUsesCacheUntilRevisionChanges'
```

## 7. Lỗi gặp phải và cách xử lý

| Vấn đề | Xử lý |
|---|---|
| `sudo` cần mật khẩu, dán nhiều dòng thì các dòng sau bị coi là mật khẩu | Chạy phần cần root bằng `wsl -d Ubuntu -u root -- bash -c "..."` từ PowerShell |
| `npm` bản Windows (`/mnt/c/Program Files/nodejs`) lẫn trong `PATH` của WSL | Cài Node bằng nvm; nvm đặt Node Linux trước trong `PATH` |
| `npm install` sửa `frontend/package-lock.json` | `git checkout frontend/package-lock.json` sau khi build |
| Gọi `wsl.exe` từ Git Bash: `$VAR` bị Git Bash thay trước | Dùng đường dẫn tuyệt đối, hoặc chạy script `.sh` |
| `build/` không nằm trong `.gitignore` của upstream | Không `git add` thư mục này |
