package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type TrackInfo struct {
	Name   string
	Artist string
}

func getSpotifyPlaylist(client *spotify.Client, playlistID spotify.ID) ([]TrackInfo, error) {
	playlist, err := client.GetPlaylist(context.Background(), playlistID)
	if err != nil {
		return nil, err
	}

	var tracks []TrackInfo
	for _, item := range playlist.Tracks.Tracks {
		track := item.Track
		tracks = append(tracks, TrackInfo{Name: track.Name, Artist: track.Artists[0].Name})
	}

	return tracks, nil
}

func searchYouTubeVideo(service *youtube.Service, query string) (string, error) {
	call := service.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(1)
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(response.Items) == 0 {
		return "", fmt.Errorf("no results found for query: %s", query)
	}
	return response.Items[0].Id.VideoId, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if spotifyClientID == "" || spotifyClientSecret == "" {
		log.Fatal("Missing SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET environment variable")
	}

	spotifyConfig := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifyClientSecret,
		TokenURL:     "https://accounts.spotify.com/api/token",
	}

	spotifyHttpClient := spotifyConfig.Client(context.Background())
	spotifyClient := spotify.New(spotifyHttpClient)
	playlistID := spotify.ID("2nSHh0BiEoRjfOAF5HXLu9")

	tracks, err := getSpotifyPlaylist(spotifyClient, playlistID)
	if err != nil {
		log.Fatalf("Error retrieving Spotify playlist: %v", err)
	}

	ytCredentialsJSON := []byte(os.Getenv("YOUTUBE_CREDENTIALS_JSON"))
	if len(ytCredentialsJSON) == 0 {
		log.Fatal("Missing YOUTUBE_CREDENTIALS_JSON environment variable")
	}

	ytConfig, err := google.ConfigFromJSON(ytCredentialsJSON, youtube.YoutubeScope)
	if err != nil {
		log.Fatalf("Error loading YouTube credentials: %v", err)
	}

	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(ytConfig)
		saveToken(tokenFile, token)
	}

	ytClient := ytConfig.Client(context.Background(), token)
	ytService, err := youtube.NewService(context.Background(), option.WithHTTPClient(ytClient))
	if err != nil {
		log.Fatalf("Error creating YouTube service: %v", err)
	}

	videoIds := make([]string, 0)

	for _, track := range tracks {
		query := fmt.Sprintf("%s %s", track.Name, track.Artist)
		videoID, err := searchYouTubeVideo(ytService, query)
		if err != nil {
			log.Printf("Error searching YouTube for %s by %s: %v", track.Name, track.Artist, err)
			continue
		}
		videoIds = append(videoIds, videoID)
	}
	for _, videoID := range videoIds {
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		cmd := exec.Command("youtube-dl", videoURL)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to download video: %v\n%s", err, output)
		}
	}
	fmt.Println("YouTube playlist created successfully!")

}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to create token file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	decodedURL, err := url.QueryUnescape(authCode)
	if err != nil {
		log.Fatalf("Unable to decode authorization code: %v", err)
	}

	fmt.Printf("Decoded URL: %s\n", decodedURL)

	parsedURL, err := url.Parse(decodedURL)
	if err != nil {
		log.Fatalf("Unable to parse URL: %v", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		log.Fatalf("Authorization code not found in URL")
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}
