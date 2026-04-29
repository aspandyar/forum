package google

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func FetchUserInfo(clientID, clientSecret, redirectURI, code string) (UserInfo, error) {
	tokenURL := "https://accounts.google.com/o/oauth2/token"
	data := fmt.Sprintf("code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code", code, clientID, clientSecret, redirectURI)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	var tokenResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return UserInfo{}, err
	}
	accessToken := tokenResponse["access_token"].(string)

	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, _ := http.NewRequest("GET", userInfoURL, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)

	userInfoResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer userInfoResp.Body.Close()

	var info UserInfo
	if err := json.NewDecoder(userInfoResp.Body).Decode(&info); err != nil {
		return UserInfo{}, err
	}
	return info, nil
}
