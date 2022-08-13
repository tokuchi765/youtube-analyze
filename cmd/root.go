package cmd

import (
	"encoding/csv"
	"encoding/json"
	"io"
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
			config, _ := loadConfig(current)
			createVideoData(config.DeveloperKey, config.ChannelID, current, o.auth)
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

func createVideoData(developerKey string, channelID string, current string, option string) {

	var client *http.Client
	if option == "oauth" {
		client = getClient(youtube.YoutubeReadonlyScope, current)
	} else if option == "api" {
		client = &http.Client{
			Transport: &transport.APIKey{Key: developerKey},
		}
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// チャンネルの動画のuploadsIDを取得する
	uploadsID := getUploadsID(channelID, service)

	// チャンネルの動画のvideoIDを全て取得する
	videoIDs := getVideoIDs(uploadsID, service)

	videoDatas := getVideoDatas(videoIDs, service)

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

	if playlistsResponse.NextPageToken != "" {
		nextPageToken := playlistsResponse.NextPageToken
		for {
			nextCall := service.PlaylistItems.List([]string{"snippet", "contentDetails"}).PlaylistId(uploadsID).PageToken(nextPageToken).MaxResults(50)
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

func getVideoDatas(videoIDs []string, service *youtube.Service) (videoDatas []entity.VideoData) {
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
	return videoDatas
}

func outputCSV(videoDatas []entity.VideoData) {
	file, _ := os.OpenFile(time.Now().Format("20060102150405")+"_youtube_data.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		file.Write([]byte{0xEF, 0xBB, 0xBF})
		w := csv.NewWriter(file)
		return gocsv.NewSafeCSVWriter(w)
	})

	gocsv.MarshalFile(videoDatas, file)

	file.Close()
}
