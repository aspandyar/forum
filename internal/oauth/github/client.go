package github

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func AccessToken(clientID, clientSecret, code string) (string, error) {
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}

	requestJSON, _ := json.Marshal(requestBodyMap)
	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respbody, _ := ioutil.ReadAll(resp.Body)

	type accessTokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	var ghresp accessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	return ghresp.AccessToken, nil
}

func UserData(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respbody, _ := ioutil.ReadAll(resp.Body)
	return string(respbody), nil
}
