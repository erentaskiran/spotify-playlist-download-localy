# Spotify playlist download locally

This project is a Go application that leverages the Spotify and YouTube APIs to manage music and video content. Users can search for music on Spotify, create playlists, and upload and update videos on YouTube.

## Getting Started

Follow these instructions to get the project up and running on your local machine.

### Prerequisites

- Go (version 1.x)
- A Spotify Developer account and a YouTube Developer account

### Installation

1. Clone this repository.
2. Fill in the `.env` file with your own Spotify and YouTube API keys. See `.envExample` for an example.

    ```properties
    SPOTIFY_CLIENT_ID = <your_spotify_client_id>
    SPOTIFY_CLIENT_SECRET = <your_spotify_client_secret>
    YOUTUBE_CREDENTIALS_JSON = <your_youtube_credentials_json>
    ```

3. Install dependencies:

    ```sh
    go mod tidy
    ```

4. Run the application:

    ```sh
    go run main.go
    ```

## Usage

Once the application is running, you can perform various operations on Spotify and YouTube. For example, you can create a playlist on Spotify or upload a video to YouTube.
