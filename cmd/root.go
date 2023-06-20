package cmd

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	"github.com/tokuchi765/youtube-analyze/entity"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

// NewCmdRoot ルートコマンドを生成します
func NewCmdRoot() *cobra.Command {
	type Options struct {
		auth string `validate:"alphanum"`
	}

	var (
		o = &Options{}
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "YouTube data csv output",
		Run: func(cmd *cobra.Command, args []string) {
			current, _ := os.Getwd()
			createReportData(current, args[0], args[1])
		},
	}

	cmd.Flags().StringVarP(&o.auth, "auth", "a", "oauth", "Choose authentication option from (api,oauth).")

	return cmd
}

// Execute コマンドライン実行
func Execute() {
	cmd := NewCmdRoot()
	cmd.SetOutput(os.Stdout)
	if err := cmd.Execute(); err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err)
		os.Exit(1)
	}
}

type config struct {
	ChannelID string `json:"channelId"`
}

func loadConfig(current string) (*config, error) {
	f, err := os.Open(current + "/config.json")
	if err != nil {
		log.Fatal("loadConfig os.Open err:", err)
		return nil, err
	}
	defer f.Close()

	var cfg config
	err = json.NewDecoder(f).Decode(&cfg)
	return &cfg, err
}

func createReportData(current string, startDate string, endDate string) {
	ctx, tok, config := getCtxTok(youtube.YoutubeReadonlyScope, current)

	client := config.Client(ctx, tok)

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	userConfig, _ := loadConfig(current)

	// チャンネルの動画のuploadsIDを取得する
	uploadsID := getUploadsID(userConfig.ChannelID, service)

	// チャンネルの動画のvideoIDを全て取得する
	videoIDs := getVideoIDs(uploadsID, service)

	videoIDsStr := strings.Join(videoIDs, ",")

	youtubeanalyticsService, _ := youtubeanalytics.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, tok)))
	call := youtubeanalyticsService.Reports.Query().Ids("channel==MINE").StartDate(startDate).EndDate(endDate).Metrics("views,estimatedMinutesWatched,averageViewDuration,comments,likes,dislikes,subscribersGained,subscribersLost").Filters("video==" + videoIDsStr).Dimensions("video")

	resp, err := call.Do()
	if err != nil {
		log.Fatalf("Error call YouTube API: %v", err)
	}

	var videoDatas []entity.VideoData
	for _, videoID := range videoIDs {
		call := service.Videos.List([]string{"snippet", "statistics"}).Id(videoID)
		response, err := call.Do()
		if err != nil {
			log.Fatalf("Error call YouTube API: %v", err)
		}

		item := response.Items[0]
		videoData := entity.VideoData{
			VideoID:            videoID,
			Title:              item.Snippet.Title,
			TotalViewCount:     item.Statistics.ViewCount,
			TotalLikeCount:     item.Statistics.LikeCount,
			TotalDislikeCount:  item.Statistics.DislikeCount,
			TotalFavoriteCount: item.Statistics.FavoriteCount,
			TotalCommentCount:  item.Statistics.CommentCount,
			PublishedAt:        item.Snippet.PublishedAt,
		}

		analytics := searchNestedArray(resp.Rows, videoID)

		if analytics != nil {
			videoData.ViewCount = analytics[1].(float64)
			videoData.EstimatedMinutesWatched = analytics[2].(float64)
			videoData.AverageViewDuration = analytics[3].(float64)
			videoData.Comments = analytics[4].(float64)
			videoData.Likes = analytics[5].(float64)
			videoData.Dislikes = analytics[6].(float64)
			videoData.SubscribersGained = analytics[7].(float64)
			videoData.SubscribersLost = analytics[8].(float64)
		}

		videoDatas = append(videoDatas, videoData)
	}

	outputCSV(videoDatas, startDate, endDate)
}

func searchNestedArray(arr [][]interface{}, target string) []interface{} {
	for _, subArr := range arr {
		if subArr[0].(string) == target {
			return subArr
		}
	}
	return nil
}

func getUploadsID(channelID string, service *youtube.Service) string {
	channelsCall := service.Channels.List([]string{"contentDetails"}).Id(channelID)
	channelsResponse, err := channelsCall.Do()
	if err != nil {
		log.Fatalf("Error call YouTube API: %v", err)
	}
	channelsItem := channelsResponse.Items[0]
	return channelsItem.ContentDetails.RelatedPlaylists.Uploads
}

func getVideoIDs(uploadsID string, service *youtube.Service) (videoIDs []string) {
	playlistsCall := service.PlaylistItems.List([]string{"contentDetails"}).PlaylistId(uploadsID).MaxResults(50)

	playlistsResponse, err := playlistsCall.Do()
	if err != nil {
		log.Fatalf("Error call YouTube API: %v", err)
	}

	for _, item := range playlistsResponse.Items {
		videoIDs = append(videoIDs, item.ContentDetails.VideoId)
	}

	if playlistsResponse.NextPageToken != "" {
		nextPageToken := playlistsResponse.NextPageToken
		for {
			nextCall := service.PlaylistItems.List([]string{"contentDetails"}).PlaylistId(uploadsID).PageToken(nextPageToken).MaxResults(50)
			nextResponse, err := nextCall.Do()
			if err != nil {
				log.Fatalf("Error call YouTube API: %v", err)
			}
			for _, nextItem := range nextResponse.Items {
				videoIDs = append(videoIDs, nextItem.ContentDetails.VideoId)
			}
			nextPageToken = nextResponse.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
	}

	return videoIDs
}

func outputCSV(videoDatas []entity.VideoData, startDate string, endDate string) {
	file, _ := os.OpenFile(time.Now().Format("20060102150405")+"_youtube_data"+"("+startDate+"_"+endDate+")"+".csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		file.Write([]byte{0xEF, 0xBB, 0xBF})
		w := csv.NewWriter(file)
		return gocsv.NewSafeCSVWriter(w)
	})

	gocsv.MarshalFile(videoDatas, file)

	file.Close()
}
