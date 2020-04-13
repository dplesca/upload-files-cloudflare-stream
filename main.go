package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	tus "github.com/eventials/go-tus"
)

// CloudflareConfig holds information necessary
// for the upload process
type CloudflareConfig struct {
	AccountID string `toml:"account_id"`
	Email     string `toml:"email"`
	APIKey    string `toml:"api_key"`
}

func main() {

	var cc CloudflareConfig
	var configFilename, videoFile string
	var videoID int
	var chunkSize int64

	flag.StringVar(&configFilename, "config", "./config.toml", "path for config file")
	flag.StringVar(&videoFile, "video", "./videofile.mp4", "path for video file")
	flag.IntVar(&videoID, "id", 0, "ivm video id for metadata store")
	flag.Int64Var(&chunkSize, "chunksize", 5, "ivm video id for metadata store")
	flag.Parse()

	if _, err := toml.DecodeFile(configFilename, &cc); err != nil {
		log.Fatal("config error: ", err)
	}
	log.Println("Config file", configFilename, "read succesfully...")

	f, err := os.Open(videoFile)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	fi, err := f.Stat()

	if err != nil {
		log.Fatal("erorr while stat file", err)
	}

	// var metadata tus.Metadata
	metadata := map[string]string{
		"filename": fmt.Sprintf("%d.mp4", videoID),
		"name":     fmt.Sprintf("%d.mp4", videoID),
		"video_id": fmt.Sprintf("%d", videoID),
	}

	fingerprint := fmt.Sprintf("%s-%d-%s", fi.Name(), fi.Size(), fi.ModTime())

	headers := make(http.Header)
	headers.Add("X-Auth-Email", cc.Email)
	headers.Add("X-Auth-Key", cc.APIKey)

	config := &tus.Config{
		ChunkSize:           chunkSize * 1024 * 1024, // Cloudflare Stream requires a minimum chunk size of 5MB.
		Resume:              false,
		OverridePatchMethod: false,
		Store:               nil,
		Header:              headers,
	}

	log.Println("start upload for file", videoFile)
	log.Println("upload information:", fingerprint, metadata)

	client, _ := tus.NewClient("https://api.cloudflare.com/client/v4/accounts/"+cc.AccountID+"/media", config)

	upload := tus.NewUpload(f, fi.Size(), metadata, fingerprint)

	uploader, _ := client.CreateUpload(upload)

	err = uploader.Upload()
	if err != nil {
		log.Fatal("Upload finished with error:", err)
	}
	log.Println("upload done for file", videoFile, uploadIDFromURL(uploader.Url()))
}

func uploadIDFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
