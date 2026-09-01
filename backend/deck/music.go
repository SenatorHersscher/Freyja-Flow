package deck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 2 * time.Second}

type NowPlayingInfo struct {
	Active   bool   `json:"active"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	CoverURL string `json:"cover_url"`
	Playing  bool   `json:"playing"`
}

// GetYTMusicStatus checks if the local Freyja Music Engine is playing
func GetYTMusicStatus() NowPlayingInfo {
	resp, err := httpClient.Get("http://localhost:8080/api/player/status")
	if err != nil || resp.StatusCode != http.StatusOK {
		return NowPlayingInfo{Active: false}
	}
	defer resp.Body.Close()

	var info NowPlayingInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err == nil {
		info.Active = true
		return info
	}
	return NowPlayingInfo{Active: false}
}

// ControlYTMusic sends a playback command to YTMusic_Engine
func ControlYTMusic(action string) error {
	url := fmt.Sprintf("http://localhost:8080/api/player/%s", action)
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
