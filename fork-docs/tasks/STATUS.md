# Trạng thái công việc

`[ ]` chưa làm · `[~]` đang làm · `[x]` xong · `[?]` chờ chủ dự án (xem QUESTIONS.md)

Chọn task **chưa xong có số nhỏ nhất** mà mọi task phụ thuộc đã `[x]`. Task có ghi *(Chủ dự án ...)* cần người làm bước đó trước; nếu chưa có thì làm task khác và nhắc trong báo cáo.

**MVP = xong T0.1 → T3.5.** Phase 7 chỉ làm khi được yêu cầu.

| | Task | Phụ thuộc | File đặc tả | Ghi chú |
|---|---|---|---|---|
| [x] | **T0.1** Tạo fork và môi trường làm việc | không | [phase-0-2.md](phase-0-2.md) | 2026-11-27. AGENTS.md bị .gitignore upstream bỏ qua → thêm bằng `git add -f`, không sửa .gitignore. WSL là Ubuntu 26.04. Commit dùng email noreply GitHub. |
| [x] | **T0.2** Build được từ mã nguồn | T0.1 | [phase-0-2.md](phase-0-2.md) | 2026-11-27. Không chạy được `1panel-core version` khi chưa cài (panic thiếu /usr/local/bin/1pctl, phiên bản nằm trong DB chứ không trong binary), xem BUILD.md mục 4. arm64 ở `build/arm64/`. Agent có 2 test fail sẵn (TLS). |
| [x] | **T0.3** Phân tích gói phát hành và luồng cài đặt của upstream | T0.1 | [phase-0-2.md](phase-0-2.md) | 2026-11-27. Gói chính thức build với `-tags=xpack` (XPanel không dùng). Nâng cấp trong panel không kiểm tra checksum. Phiên bản chỉ nằm trong `ORIGINAL_VERSION` của 1pctl. Xem mục 4 của RELEASE-FORMAT.md. |
| [x] | **T1.1** CI kiểm tra build | T0.2 | [phase-0-2.md](phase-0-2.md) | 2026-11-27. Run 45900232237 xanh, 5 phút 3 giây. Bỏ qua 2 test agent fail sẵn (TLS) bằng `-skip`, không sửa test. Đã tắt tay 3 workflow upstream (Gitee, Issue Translator, PR Handling). |
| [ ] | **T1.2** Máy chủ phát hành và workflow release | T0.3, T1.1 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T1.3** URL máy chủ đọc từ cấu hình [fork] | T0.3 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T1.4** Installer riêng | T1.2, T1.3 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T1.5** Nâng cấp qua nút trong panel | T1.4 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T2.1** Kiểm kê và script đổi thương hiệu | T0.2 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T2.2** Ẩn quảng cáo bản Pro [fork] | T2.1 | [phase-0-2.md](phase-0-2.md) | |
| [ ] | **T3.1** Dịch giao diện frontend | T0.2 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T3.2** Đăng ký ngôn ngữ `vi` [fork] | T3.1 (có thể làm song song với file dịch còn dở) | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T3.3** Dịch thông báo backend | T3.2 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T3.4** Script kiểm tra key dịch | T3.1 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T3.5** Gửi bản dịch lên upstream | T3.1–T3.4 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T4.1** Dựng trang tài liệu riêng | T2.1 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T4.2** Dịch tài liệu ưu tiên | T4.1 | [phase-3-4.md](phase-3-4.md) | |
| [ ] | **T5.1** Mirror App Store | T1.3, T1.4 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T5.2** Mô tả ứng dụng tiếng Việt và ứng dụng riêng | T5.1, T3.2 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T6.1** Quy trình đồng bộ upstream | T1.5, T2.1 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T6.2** Test tự động trên VM | T1.4 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T6.3** Điều tra KVM helper | T0.2 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T6.4** Hướng dẫn chuyển từ 1Panel gốc sang XPanel | T1.5 | [phase-5-7.md](phase-5-7.md) | |
| [ ] | **T7.x** Tính năng riêng (xem phase-5-7.md) | chủ dự án yêu cầu | [phase-5-7.md](phase-5-7.md) | |
