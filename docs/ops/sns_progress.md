# docs/mvp 更新時の SNS 自動投稿

`docs/mvp/` 配下の Markdown が更新されたときに、GitHub Actions で要約を作成し、X と Threads へ同時投稿します。

## 目的

この運用の目的はバズではなく、以下の開発ログ運用です。

- 作品の存在を継続的に知らせる
- 早期フィードバックを得る
- 公開ログとして開発履歴を残す
- 開発モチベーション維持
- 将来のユーザー候補を少しずつ増やす

## 追加したもの

- Workflow: `.github/workflows/mvp-sns-summary.yml`
- 投稿文生成: `scripts/social/build_mvp_summary.py`
- Threads 投稿: `scripts/social/post_to_threads.sh`

## 必要な GitHub Secrets

### X

- `X_API_KEY`
- `X_API_SECRET`
- `X_ACCESS_TOKEN`
- `X_ACCESS_TOKEN_SECRET`

### Threads

- `THREADS_USER_ID`
- `THREADS_ACCESS_TOKEN`

## 投稿フロー

1. push で `docs/mvp/**` が変更される
2. 変更ファイルと差分の追加行から「まとめ」を自動生成
3. X と Threads に投稿（Secrets が揃っている場合のみ）
4. Secrets 不足時は workflow のログに通知だけ出して終了

## 補足

- 投稿文にはコミット URL が含まれます。
- 文字数超過時は末尾を省略して投稿できる長さに丸めます。
- 手動実行は `workflow_dispatch` から可能です。
