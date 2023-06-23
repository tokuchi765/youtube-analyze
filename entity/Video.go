package entity

// VideoData 動画情報を保持する
type VideoData struct {
	VideoID                 string  `csv:"ビデオID"`
	Title                   string  `csv:"タイトル"`
	TotalViewCount          uint64  `csv:"総再生数"`
	TotalLikeCount          uint64  `csv:"総高評価数"`
	TotalDislikeCount       uint64  `csv:"総低評価数"`
	TotalFavoriteCount      uint64  `csv:"総お気に入り数"`
	TotalCommentCount       uint64  `csv:"総コメント数"`
	PublishedAt             string  `csv:"公開日時"`
	ViewCount               float64 `csv:"期間内再生数"`
	EstimatedMinutesWatched float64 `csv:"期間内再生時間(分)"`
	AverageViewDuration     float64 `csv:"期間内平均視聴時間(秒)"`
	Comments                float64 `csv:"期間内コメント数"`
	Likes                   float64 `csv:"期間内高評価数"`
	Dislikes                float64 `csv:"期間内低評価数"`
	SubscribersGained       float64 `csv:"期間内登録回数"`
	SubscribersLost         float64 `csv:"期間内登録解除回数"`
}
