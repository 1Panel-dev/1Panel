# Quy trình cho worker

## Mỗi task
1. Đọc hết mục task: **Mục tiêu**, **Phụ thuộc**, **Việc cần làm**, **Nghiệm thu**.
2. `STATUS.md`: đổi `[ ]` → `[~]`.
3. Làm đúng đặc tả. Tên file, đường dẫn, tên biến cấu hình phải giống y đặc tả.
4. Chạy kiểm tra chung trong `AGENTS.md` và mọi mục **Nghiệm thu**.
5. `STATUS.md`: đổi → `[x]`, ghi 1 dòng (ngày, điểm khác đặc tả).
6. Commit: `git commit -m "T<số>: <tiêu đề>"` (patch vào file upstream thì tách commit `[fork] ...` riêng, xem AGENTS.md quy tắc 3), rồi **`git push`**. Báo cáo ghi rõ đã push.
7. Báo cáo theo mẫu bên dưới rồi **dừng chờ**.

## Khi vướng
Đặc tả không khớp với code upstream thực tế (upstream đổi cấu trúc, file không còn...), hoặc cần quyết định chưa có trong đặc tả: ghi vào `fork-docs/tasks/QUESTIONS.md` theo mẫu, đổi task thành `[?]`, dừng. **Không tự chọn hướng khác.**

```markdown
## T2.2: <tiêu đề>
- Vấn đề: ...
- Phương án A / B: ...
- Đề xuất: ...
```

## Mẫu báo cáo
```
T<số> xong: <tiêu đề>
- Đã làm: ...
- Kiểm tra: <lệnh + kết quả>
- Khác đặc tả: <không / ...>
- Chưa thử được: ...
- Task tiếp theo: T<số>
```

## Không làm
- Không refactor, đổi format hay "dọn dẹp" code upstream.
- Không nâng phiên bản thư viện của upstream.
- Không chạy `npm audit fix`, `go get -u` trên toàn repo.
- Không tắt test của upstream để CI xanh; test hỏng thì ghi vào QUESTIONS.md.
