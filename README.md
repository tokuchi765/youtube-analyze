# youtube-analyze

YouTubeチャンネルのデータをまとめてCSVに出力します

# 前提
1. GCP にアカウント登録済
2. 以下のAPIライブラリを有効にしたプロジェクトを用意する
![image](https://github.com/tokuchi765/youtube-analyze/assets/55987154/92da4f4b-a68e-4d97-88da-7952e6c09372)

# 使い方

1. exeファイルを任意のフォルダに置きます
2. 上記のフォルダに `client_secrets.json.example` ファイルをコピーします
3. `client_secrets.json` にリネームします
4. `client_id` をGCPで登録したクライアントIDに書き換えます
5. `client_secret` をGCPで登録したシークレットキーに書き換えます
6. 上記のフォルダに `config.json.example` ファイルをコピーします
7. `config.json` にリネームします
8. `channelId` を集計したいチャンネルに書き換えます
9. exeファイルを実行します
10. 実行の際は、第一引数に集計開始日、第二引数に集計終了日を入力してください ※日付フォーマットは `yyyy-mm-dd`<br>
`例：./main.exe 2023-06-01 2023-06-30`