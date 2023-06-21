package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	clientName    = "WEB"
	clientVersion = "2.20210721.00.00"
	ytApiUrl      = "https://www.youtube.com/youtubei/v1/player"
)

type ClientData struct {
	ClientName    string `json:"clientName"`
	CLientVersion string `json:"clientVersion"`
}

type ContextData struct {
	Client ClientData `json:"client"`
}

type PlayerRequest struct {
	Context        ContextData `json:"context"`
	VideoID        string      `json:"videoId"`
	RacyCheckOk    bool        `json:"racyCheckOk"`
	ContentCheckOk bool        `json:"contentCheckOk"`
}

type Video struct {
	Url string `json:"url"`
}

type StreamingData struct {
	ExpiresInSeconds string  `json:"expiresInSeconds"`
	Formats          []Video `json:"formats"`
}

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
}

func GetStreamUrl(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	videoId := u.Query().Get("v")
	playerResponse, err := GetVideoData(videoId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(playerResponse.StreamingData.Formats)

	return playerResponse.StreamingData.Formats[0].Url
}

func GetVideoData(videoId string) (PlayerResponse, error) {
	requestData := PlayerRequest{
		Context: ContextData{
			Client: ClientData{
				ClientName:    clientName,
				CLientVersion: clientVersion,
			},
		},
		VideoID:        videoId,
		RacyCheckOk:    false,
		ContentCheckOk: false,
	}
	jsonData, err := json.Marshal(requestData)
	request, err := http.NewRequest("POST", ytApiUrl, bytes.NewBuffer(jsonData))
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return PlayerResponse{}, errors.New("Request failed")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PlayerResponse{}, errors.New("Can't parse response")
	}

	var response PlayerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return PlayerResponse{}, errors.New("Can't parse response")
	}

	return response, nil
}
