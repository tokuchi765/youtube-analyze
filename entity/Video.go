package entity

// VideoData 動画情報を保持する
type VideoData struct {
	VideoID       string `csv:id`
	Title         string `csv:titel`
	ViewCount     uint64 `csv:view count`
	LikeCount     uint64 `csv:like count`
	DislikeCount  uint64 `csv:dislike count`
	FavoriteCount uint64 `csv:favorite count`
	CommentCount  uint64 `csv:comment count`
	PublishedAt   string `csv:published at`
}
