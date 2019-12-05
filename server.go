package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/evalphobia/google-home-client-go/googlehome"
)

const twitchAPIKey = ""
const twitchUserID = "8095777"
const followedStreamersURL = "https://api.twitch.tv/helix/users/follows?from_id=" + twitchUserID + "&first=100"
const streamStatusURL = "https://api.twitch.tv/helix/streams"

type TwitchFollowsResponse struct {
	Data []FollowInfo `json:"data"`
}

type FollowInfo struct {
	ToID   string `json:"to_id"`
	ToName string `json:"to_name"`
}

type OnlineUsersResponse struct {
	Data []struct {
		UserName string `json:"user_name"`
	} `json:"data"`
}

func fetchTwitchFollows(client http.Client) (TwitchFollowsResponse, error) {
	req, _ := http.NewRequest("GET", followedStreamersURL, nil)
	req.Header.Set("Client-ID", twitchAPIKey)

	var twitchFollowersData TwitchFollowsResponse

	res, error := client.Do(req)
	if error != nil {
		log.Fatalln(error)
	}

	defer res.Body.Close()

	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		return twitchFollowersData, errors.New("Error reading response")
	}

	err := json.Unmarshal(body, &twitchFollowersData)
	if err != nil {
		log.Fatalln(err)
		return twitchFollowersData, errors.New("Error parsing JSON")
	}

	fmt.Printf("contents of decoded json is: %#v\r\n", twitchFollowersData)
	return twitchFollowersData, nil
}

func fetchTwitchStreamersStatus(client http.Client, twitchFollowsResponse TwitchFollowsResponse) (OnlineUsersResponse, error) {
	streamStatusRequest, error := http.NewRequest("GET", streamStatusURL, nil)
	streamStatusRequest.Header.Set("Client-ID", twitchAPIKey)

	var onlineUsersResponse OnlineUsersResponse

	q := streamStatusRequest.URL.Query()
	q.Add("first", "100")
	for _, element := range twitchFollowsResponse.Data {
		q.Add("user_id", element.ToID)
		// fmt.Fprintf(w, "Username: "+element.ToID)
	}

	streamStatusRequest.URL.RawQuery = q.Encode()

	streamStatusResponse, error := client.Do(streamStatusRequest)
	if error != nil {
		log.Fatalln(error)
	}

	defer streamStatusResponse.Body.Close()

	statusResponseBody, error := ioutil.ReadAll(streamStatusResponse.Body)
	if error != nil {
		// fmt.Fprintf(w, "Error parsing data")
	}

	err := json.Unmarshal(statusResponseBody, &onlineUsersResponse)
	if err != nil {
		log.Fatalln(err)
		return onlineUsersResponse, errors.New("Error parsing JSON")
	}

	return onlineUsersResponse, nil
}

func speakOnGoogleHome(text string) {
	cli, err := googlehome.NewClientWithConfig(googlehome.Config{
		Hostname: "192.168.86.32",
		Lang:     "en",
		Accent:   "us",
	})
	if err != nil {
		panic(err)
	}

	// Speak text on Google Home.
	cli.Notify(text)
}

func twitchChannelList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Requesting")

	client := http.Client{}

	twitchFollowsResponse, error := fetchTwitchFollows(client)
	if error != nil {
		// return a 500 here
		return
	}

	onlineUsersResponse, error := fetchTwitchStreamersStatus(client, twitchFollowsResponse)
	if error != nil {
		// return a 500 here
		return
	}

	onlineUsersString := ""
	for index, user := range onlineUsersResponse.Data {
		onlineUsersString += strconv.Itoa(index+1) + " " + user.UserName + ", "
	}

	if len(onlineUsersResponse.Data) == 1 {
		onlineUsersString += "is online"
	} else {
		onlineUsersString += "are online"
	}

	speakOnGoogleHome(onlineUsersString)

	fmt.Fprintf(w, "Online users: "+onlineUsersString)
}

func main() {
	http.HandleFunc("/twitch-channel-list", twitchChannelList)
	http.ListenAndServe(":80", nil)
}
