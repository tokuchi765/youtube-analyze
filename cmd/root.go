package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"
	"github.com/tokuchi765/youtube-analyze/entity"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

// RootCmd is root command
var RootCmd = &cobra.Command{
	Use:   "analyze",
	Short: "YouTube data csv output",
	Run: func(cmd *cobra.Command, args []string) {
		current, _ := os.Getwd()
		config, _ := loadConfig(current)
		createVideoData(config.DeveloperKey, config.ChannelID)
	},
}

type config struct {
	DeveloperKey string `json:"developerKey"`
	ChannelID    string `json:"channelId"`
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

func createVideoData(developerKey string, channelID string) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// チャンネルの動画のuploadsIDを取得する
	uploadsID := getUploadsID(channelID, service)

	// チャンネルの動画のvideoIDを全て取得する
	videoIDs := getVideoIDs(uploadsID, service)

	var videoDatas []entity.VideoData
	for _, videoID := range videoIDs {
		call := service.Videos.List([]string{"snippet", "statistics"}).Id(videoID)
		response, err := call.Do()
		if err != nil {
			log.Fatalf("Error call YouTube API: %v", err)
		}

		item := response.Items[0]
		videoData := entity.VideoData{
			VideoID:       videoID,
			Title:         item.Snippet.Title,
			ViewCount:     item.Statistics.ViewCount,
			LikeCount:     item.Statistics.LikeCount,
			DislikeCount:  item.Statistics.DislikeCount,
			FavoriteCount: item.Statistics.FavoriteCount,
			CommentCount:  item.Statistics.CommentCount,
			PublishedAt:   item.Snippet.PublishedAt,
		}
		videoDatas = append(videoDatas, videoData)
	}

	outputCSV(videoDatas)
}

func getUploadsID(channelID string, service *youtube.Service) string {
	channelsCall := service.Channels.List([]string{"snippet", "contentDetails", "statistics"}).Id(channelID)
	channelsResponse, err := channelsCall.Do()
	if err != nil {
		log.Fatalf("Error call YouTube API: %v", err)
	}
	channelsItem := channelsResponse.Items[0]
	return channelsItem.ContentDetails.RelatedPlaylists.Uploads
}

func getVideoIDs(uploadsID string, service *youtube.Service) (videoIDs []string) {
	playlistsCall := service.PlaylistItems.List([]string{"snippet", "contentDetails"}).PlaylistId(uploadsID).MaxResults(50)
	playlistsResponse, err := playlistsCall.Do()
	if err != nil {
		log.Fatalf("Error call YouTube API: %v", err)
	}

	for _, item := range playlistsResponse.Items {
		videoIDs = append(videoIDs, item.ContentDetails.VideoId)
	}

	return videoIDs
}

func outputCSV(videoDatas []entity.VideoData) {
	file, _ := os.OpenFile(time.Now().Format("20060102150405")+"_youtube_data.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	gocsv.MarshalFile(videoDatas, file)

	file.Close()
}
