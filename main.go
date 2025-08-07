package main

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"main/ynison"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const SignSalt = "XGRlBW9FXlekgbPrRHuSiA"

type CurrentTrackInfo struct {
	Paused     bool            `json:"paused,omitempty"`
	DurationMs string          `json:"duration_ms,omitempty"`
	ProgressMs string          `json:"progress_ms,omitempty"`
	EntityID   string          `json:"entity_id,omitempty"`
	EntityType string          `json:"entity_type,omitempty"`
	Track      YandexTrackInfo `json:"track"`
}

type YandexTrackInfo struct {
	TrackID     int    `json:"track_id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	ImageURL    string `json:"img"`
	Duration    int    `json:"duration"`
	Minutes     int    `json:"minutes"`
	Seconds     int    `json:"seconds"`
	AlbumID     int    `json:"album_id"`
	DownloadURL string `json:"download_link"`
}

type SoundCloudTrackInfo struct {
	ID           int    `json:"id"`
	PermalinkUrl string `json:"permalink_url"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	ArtworkURL   string `json:"artwork_url"`
	Duration     int    `json:"duration"`
	Minutes      int    `json:"minutes"`
	Seconds      int    `json:"seconds"`
}

type DownloadXMLData struct {
	Host string `xml:"host"`
	Path string `xml:"path"`
	TS   string `xml:"ts"`
	S    string `xml:"s"`
}

type TempTrackInfo struct {
	Result []struct {
		Title      string `json:"title"`
		DurationMs int    `json:"durationMs"`
		CoverURI   string `json:"coverUri"`
		Artists    []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Albums []struct {
			ID int `json:"id"`
		} `json:"albums"`
	} `json:"result"`
}

type TempDownloadInfo struct {
	Result []struct {
		DownloadLink string `json:"downloadInfoUrl"`
	} `json:"result"`
}

var client = new(http.Client)

func main() {
	mTLSConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	mTLSConfig.MinVersion = tls.VersionTLS11
	mTLSConfig.MaxVersion = tls.VersionTLS13

	tr := &http.Transport{
		TLSClientConfig: mTLSConfig,
	}

	client.Transport = tr
	client.Timeout = 5 * time.Second
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.GET("/get_current_track_beta", GetCurrentTrackYandex)
	r.GET("/get_current_track_soundcloud", GetCurrentTrackSoundcloud)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}

func GetCurrentTrackYandex(c *gin.Context) {
	token := c.Request.Header.Get("ya-token")
	tokenHeader := c.Request.Header.Get("Authorization")

	if token == "" && tokenHeader == "" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{
			"place": "early",
			"error": "token not found",
		})
		return
	}

	if token != "" {
		tokenHeader = "OAuth " + token
	}

	done := make(chan ynison.PutYnisonStateResponse, 1)
	y := ynison.NewClient(tokenHeader)
	defer y.Close()

	y.OnMessage(func(pysr ynison.PutYnisonStateResponse) {
		done <- pysr
	})

	err := y.Connect()
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{
			"place": "ynison connect",
			"error": err.Error(),
		})
		return
	}

	select {
	case <-time.After(10 * time.Second):
		c.IndentedJSON(http.StatusGatewayTimeout, gin.H{
			"place": "context.Done",
			"error": "context.Done",
		})
	case <-c.Request.Context().Done():
		c.IndentedJSON(http.StatusGatewayTimeout, gin.H{
			"place": "10 seconds timeout",
			"error": "Failed to retrieve data",
		})
	case data := <-done:
		if len(data.PlayerState.PlayerQueue.PlayableList) <= 0 {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "PlayerQueue information missing"})
			return
		}

		index := data.PlayerState.PlayerQueue.CurrentPlayableIndex
		trackID := data.PlayerState.PlayerQueue.PlayableList[index].PlayableID
		track, err := GetTrackInfoYandex(trackID, tokenHeader)

		if err != nil {
			c.IndentedJSON(http.StatusTeapot, gin.H{
				"place": "get track",
				"error": err.Error(),
			})
			return
		}

		var result = new(CurrentTrackInfo)
		result.Paused = data.PlayerState.Status.Paused
		result.DurationMs = data.PlayerState.Status.DurationMs
		result.ProgressMs = data.PlayerState.Status.ProgressMs
		result.EntityID = data.PlayerState.PlayerQueue.EntityID
		result.EntityType = data.PlayerState.PlayerQueue.EntityType
		result.Track = *track
		c.IndentedJSON(http.StatusOK, result)
	}
}

func GetTrackInfoYandex(trackId string, header string) (*YandexTrackInfo, error) {
	trackInfo := new(TempTrackInfo)
	err := LoadInfo("https://api.music.yandex.net/tracks/"+trackId, header, trackInfo)
	if err != nil {
		return nil, err
	}

	downloadInfo := new(TempDownloadInfo)
	err = LoadInfo("https://api.music.yandex.net/tracks/"+trackId+"/download-info", header, downloadInfo)
	if err != nil {
		return nil, err
	}

	downloadXmlResponse, err := SendRequest(downloadInfo.Result[0].DownloadLink, header)
	if downloadXmlResponse == nil {
		return nil, err
	}

	defer downloadXmlResponse.Body.Close()
	if downloadXmlResponse.StatusCode != 200 {
		return nil, err
	}

	downloadXmlBytes, _ := io.ReadAll(downloadXmlResponse.Body)
	downloadXml := string(downloadXmlBytes)
	downloadUrl, err := BuildDirectLink([]byte(downloadXml))

	if err != nil {
		return nil, err
	}

	var artists string
	for i, artist := range trackInfo.Result[0].Artists {
		artists += artist.Name
		if i != len(trackInfo.Result[0].Artists)-1 {
			artists += ", "
		}
	}

	if len(trackInfo.Result[0].Artists) == 0 {
		artists = "Артист не найден"
	}

	result := new(YandexTrackInfo)
	result.Title = trackInfo.Result[0].Title
	result.Artist = artists
	result.ImageURL = "https://" + strings.Replace(trackInfo.Result[0].CoverURI, "%%", "1000x1000", 1)
	result.Duration = trackInfo.Result[0].DurationMs / 1000
	result.Minutes = result.Duration / 60
	result.Seconds = result.Duration % 60
	result.DownloadURL = downloadUrl

	result.TrackID, _ = strconv.Atoi(trackId)
	if result.TrackID == 0 {
		result.TrackID = -1
	}

	result.AlbumID = -1
	if len(trackInfo.Result[0].Albums) > 0 {
		result.AlbumID = trackInfo.Result[0].Albums[0].ID
	}

	if result.ImageURL == "https://" {
		result.ImageURL = "https://raw.githubusercontent.com/MIPOHBOPOHIH/MIPOHSITE/refs/heads/main/public/images/nonetrack.png"
	}

	return result, nil
}

func BuildDirectLink(xmlBytes []byte) (string, error) {
	var data DownloadXMLData
	err := xml.Unmarshal(xmlBytes, &data)
	if err != nil {
		return "", err
	}

	trimmedPath := ""
	if len(data.Path) > 1 {
		trimmedPath = data.Path[1:]
	}

	hash := md5.Sum([]byte(SignSalt + trimmedPath + data.S))
	sign := hex.EncodeToString(hash[:])

	url := fmt.Sprintf("https://%s/get-mp3/%s/%s%s", data.Host, sign, data.TS, data.Path)
	return url, nil
}

func LoadInfo[T any](url string, header string, container *T) error {
	response, err := SendRequest(url, header)
	if response == nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return err
	}

	err = json.NewDecoder(response.Body).Decode(container)
	return nil
}

func SendRequest(url string, header string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", header)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, err
	}

	return response, nil
}

// --- SoundCloud API structures and logic ---
// практически весь код я спиздил https://github.com/es3n1n/nowplaying
// он мне разрешил
// а еще код сделал чатгпт и его никто не ревьюил, поэтому экстрапохуй
type SoundCloudAPITrack struct {
	ID           int    `json:"id"`
	PermalinkUrl string `json:"permalink_url"`
	Title        string `json:"title"`
	User         struct {
		Username string `json:"username"`
	} `json:"user"`
	PublisherMetadata struct {
		Artist string `json:"artist"`
	} `json:"publisher_metadata"`
	ArtworkURL   string `json:"artwork_url"`
	Downloadable bool   `json:"downloadable"`
	Duration     int    `json:"duration"`
	Media        struct {
		Transcodings []struct {
			URL    string `json:"url"`
			Format struct {
				Protocol string `json:"protocol"`
				MimeType string `json:"mime_type"`
			} `json:"format"`
		} `json:"transcodings"`
	} `json:"media"`
}

type SoundCloudHistoryItem struct {
	Track    SoundCloudAPITrack `json:"track"`
	PlayedAt int64              `json:"played_at"`
}

type SoundCloudHistoryResponse struct {
	Collection []SoundCloudHistoryItem `json:"collection"`
}

func getSoundCloudTrackInfo(track SoundCloudAPITrack) SoundCloudTrackInfo {
	author := track.User.Username
	if track.PublisherMetadata.Artist != "" {
		author = track.PublisherMetadata.Artist
	}
	minutes := track.Duration / 60000
	seconds := (track.Duration % 60000) / 1000
	return SoundCloudTrackInfo{
		ID:           track.ID,
		PermalinkUrl: track.PermalinkUrl,
		Title:        track.Title,
		Author:       author,
		ArtworkURL:   track.ArtworkURL,
		Duration:     track.Duration,
		Minutes:      minutes,
		Seconds:      seconds,
	}
}

func getSoundCloudHistory(scToken string) (*SoundCloudHistoryItem, error) {
	apiURL := "https://api-v2.soundcloud.com/me/play-history/tracks?limit=1&offset=0&client_id=1HxML01xkzWgtHfBreaeZfpANMe3ADjb"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+scToken)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("SoundCloud API error: %d", resp.StatusCode)
	}
	var historyResp SoundCloudHistoryResponse
	err = json.NewDecoder(resp.Body).Decode(&historyResp)
	if err != nil {
		return nil, err
	}
	if len(historyResp.Collection) == 0 {
		return nil, nil
	}
	return &historyResp.Collection[0], nil
}

func GetCurrentTrackSoundcloud(c *gin.Context) {
	scToken := c.GetHeader("sc-token")
	if scToken == "" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Необходим заголовок sc-token"})
		return
	}

	historyItem, err := getSoundCloudHistory(scToken)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if historyItem == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "История пуста"})
		return
	}
	trackInfo := getSoundCloudTrackInfo(historyItem.Track)
	c.IndentedJSON(http.StatusOK, gin.H{"track": trackInfo})
}
