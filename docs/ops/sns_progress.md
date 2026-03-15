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

> 設定場所: GitHub リポジトリの **Settings > Secrets and variables > Actions > New repository secret**

## 投稿フロー

1. `docs/mvp/**` の変更が push されたら workflow が起動
2. 変更ファイルと差分の追加行から「まとめ」を自動生成
3. X と Threads に投稿（Secrets が揃っている場合のみ）
4. Secrets 不足時は workflow のログに通知だけ出して終了

## 検知タイミング（重要）

- **投稿実行は push 時のみ** です（ローカルコミットだけでは動きません）。
- `pull_request` でも要約生成は動作し、PR の Actions ログで確認できます。
- 新規作成ファイル（例: `docs/mvp/08.md`）も `docs/mvp/**` に含まれるため検知対象です。

## 投稿内容

- 基本は「変更された `0X.md` ファイルの要約」です。
- 具体的には、差分の **追加行** からハイライトを抽出して投稿文を作ります。
- 投稿文にはコミット URL を含めます。
- 文字数超過時は末尾を省略して投稿できる長さに丸めます。
- 手動実行は `workflow_dispatch` から可能です。
