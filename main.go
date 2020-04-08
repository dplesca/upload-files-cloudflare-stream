package main

import (
	"flag"
	"log"
	"net/http"
	"os"

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

	flag.StringVar(&configFilename, "config", "./config.toml", "path for config file")
	flag.StringVar(&videoFile, "video", "./videofile.mp4", "path for video file")
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

	headers := make(http.Header)
	headers.Add("X-Auth-Email", cc.Email)
	headers.Add("X-Auth-Key", cc.APIKey)

	config := &tus.Config{
		ChunkSize:           5 * 1024 * 1024, // Cloudflare Stream requires a minimum chunk size of 5MB.
		Resume:              false,
		OverridePatchMethod: false,
		Store:               nil,
		Header:              headers,
	}

	log.Println("start upload for file", videoFile)

	client, _ := tus.NewClient("https://api.cloudflare.com/client/v4/accounts/"+cc.AccountID+"/media", config)

	upload, _ := tus.NewUploadFromFile(f)

	uploader, _ := client.CreateUpload(upload)

	err = uploader.Upload()
	if err != nil {
		log.Fatal("Upload finished with error:", err)
	}
	log.Println("upload done for file", videoFile)
}
