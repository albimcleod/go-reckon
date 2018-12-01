package goreckon

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL        = "https://identity.reckon.com"
	tokenURL       = "connect/token"
	refreshURL     = "connect/token"
	ApiEndpointURL = "https://api.reckon.com"
)

var (
	defaultSendTimeout = time.Second * 30
)

type Reckon struct {
	StoreCode    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Timeout      time.Duration
}

// NewClient will create a Reckon client with default values
func NewClient(code string, clientID string, clientSecret string, redirectURI string) *Reckon {
	return &Reckon{
		StoreCode:    code,
		Timeout:      defaultSendTimeout,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}
}

// AccessToken will get a new access token
func (v *Reckon) AccessToken() (string, string, time.Time, error) {

	u, _ := url.ParseRequestURI(baseURL)
	u.Path = tokenURL
	urlStr := fmt.Sprintf("%v", u)

	request := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", v.StoreCode, v.RedirectURI)

	fmt.Println("--------------------------")
	fmt.Println("Calling Reckon", urlStr)
	fmt.Println(string(request))

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer([]byte(request)))

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	auth := fmt.Sprintf("%s:%s", v.ClientID, v.ClientSecret)
	sEnc := b64.StdEncoding.EncodeToString([]byte(auth))

	r.Header.Add("Authorization", "Basic "+sEnc)
	fmt.Println("Authorization", "Basic "+sEnc)

	fmt.Println("--------------------------")

	res, _ := client.Do(r)

	rawResBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", "", time.Now(), fmt.Errorf("%v", string(rawResBody))
	}

	if res.StatusCode == 200 {
		resp := &TokenResponse{}
		if err := json.Unmarshal(rawResBody, resp); err != nil {
			return "", "", time.Now(), err
		}
		return resp.AccessToken, resp.RefreshToken, time.Now().Add(time.Duration(resp.ExpiresIn) * time.Millisecond), nil
	}

	fmt.Println(string(rawResBody))

	return "", "", time.Now(), fmt.Errorf("Failed to get access token: %s", res.Status)
}

// RefreshToken will get a new refresg token
func (v *Reckon) RefreshToken(refreshtoken string) (string, string, time.Time, error) {
	u, _ := url.ParseRequestURI(baseURL)
	u.Path = tokenURL
	urlStr := fmt.Sprintf("%v", u)

	request := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&redirect_uri=%s", refreshtoken, v.RedirectURI)

	fmt.Println("--------------------------")
	fmt.Println("Calling Reckon", urlStr)
	fmt.Println(string(request))

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer([]byte(request)))

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	auth := fmt.Sprintf("%s:%s", v.ClientID, v.ClientSecret)
	sEnc := b64.StdEncoding.EncodeToString([]byte(auth))

	r.Header.Add("Authorization", "Basic "+sEnc)

	fmt.Println("--------------------------")

	res, _ := client.Do(r)

	rawResBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", "", time.Now(), fmt.Errorf("%v", string(rawResBody))
	}

	if res.StatusCode == 200 {
		resp := &TokenResponse{}
		if err := json.Unmarshal(rawResBody, resp); err != nil {
			return "", "", time.Now(), err
		}
		return resp.AccessToken, resp.RefreshToken, time.Now().Add(time.Duration(resp.ExpiresIn) * time.Millisecond), nil
	}

	fmt.Println(string(rawResBody))

	return "", "", time.Now(), fmt.Errorf("Failed to get refresh token: %s", res.Status)
}

// GetCompany will return the authenticated company
func (v *Reckon) GetContacts(token string, bookId string) ([]Contact, error) {
	client := &http.Client{}
	client.CheckRedirect = checkRedirectFunc

	u, _ := url.ParseRequestURI(ApiEndpointURL)
	u.Path = fmt.Sprintf("r1/%s/contacts", bookId)
	urlStr := fmt.Sprintf("%v", u)

	r, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("urlStr", urlStr)

	r.Header = http.Header(make(map[string][]string))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+token)

	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	//	fmt.Println("GetCompany Body", string(rawResBody))

	if res.StatusCode == 200 {
		var resp []Contact

		err = json.Unmarshal(rawResBody, &resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, fmt.Errorf("Failed to get Reckon Contacts %s", res.Status)

}

// GetBooks will return the authenticated company
func (v *Reckon) GetBooks(token string) ([]Book, error) {
	client := &http.Client{}
	client.CheckRedirect = checkRedirectFunc

	u, _ := url.ParseRequestURI(ApiEndpointURL)
	u.Path = "r1/cashbooks"
	urlStr := fmt.Sprintf("%v", u)

	r, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("urlStr", urlStr)

	r.Header = http.Header(make(map[string][]string))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+token)

	fmt.Println("token", token)

	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("GetBooks Body", string(rawResBody))

	if res.StatusCode == 200 {
		var resp []Book

		err = json.Unmarshal(rawResBody, &resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, fmt.Errorf("Failed to get Reckon Contacts %s", res.Status)

}

func checkRedirectFunc(req *http.Request, via []*http.Request) error {
	if req.Header.Get("Authorization") == "" {
		req.Header.Add("Authorization", via[0].Header.Get("Authorization"))
	}
	return nil
}
