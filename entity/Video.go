package entity

// VideoData 動画情報を保持する
type VideoData struct {
	VideoID                 string  `csv:id`
	Title                   string  `csv:titel`
	TotalViewCount          uint64  `csv:total view count`
	TotalLikeCount          uint64  `csv:total like count`
	TotalDislikeCount       uint64  `csv:total dislike count`
	TotalFavoriteCount      uint64  `csv:total favorite count`
	TotalCommentCount       uint64  `csv:total comment count`
	PublishedAt             string  `csv:published at`
	ViewCount               float64 `csv:view count`
	EstimatedMinutesWatched float64 `csv:estimated minutes watched`
	AverageViewDuration     float64 `csv:average view duration`
	Comments                float64 `csv:comments`
	Likes                   float64 `csv:likes`
	Dislikes                float64 `csv:dislikes`
	SubscribersGained       float64 `csv:subscribers gained`
	SubscribersLost         float64 `csv:subscribers lost`
}
