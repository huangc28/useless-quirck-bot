#!/bin/sh
set -eu

input="$(cat)"
observed_at="2026-05-28T12:00:00Z"

case "$input" in
  *auth_required*|*登入*|*會員*)
    printf '{"status":"auth_required","answer":"","evidence":[],"observed_at":"%s","error":"browser session required"}\n' "$observed_at"
    ;;
  *worker_error*|*error*)
    printf '{"status":"error","answer":"","evidence":[],"observed_at":"%s","error":"mock worker error"}\n' "$observed_at"
    ;;
  *no_result*|*找不到*)
    printf '{"status":"no_result","answer":"找不到符合條件的資料。","evidence":[],"observed_at":"%s","error":""}\n' "$observed_at"
    ;;
  *browser_lookup*)
    printf '{"status":"ok","answer":"momo 上的義美小泡芙價格需以頁面即時顯示為準。","evidence":[{"title":"momo 商品頁","url":"https://www.momoshop.com.tw/","source":"momo","snippet":"Mock browser lookup result","observed_at":"%s"}],"observed_at":"%s","error":""}\n' "$observed_at" "$observed_at"
    ;;
  *)
    printf '{"status":"ok","answer":"義美小泡芙公開價格約 NT$59，實際價格請以來源頁面為準。","evidence":[{"title":"公開商品搜尋結果","url":"https://example.com/ime-puff","source":"mock-public","snippet":"Mock public lookup result","observed_at":"%s"}],"observed_at":"%s","error":""}\n' "$observed_at" "$observed_at"
    ;;
esac
