package gifs

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

type Giphy struct {
	token string
}

func NewGiphy(token string) *Giphy {
	return &Giphy{token: token}
}

func (g *Giphy) Gif(query string) (string, error) {
	u, err := url.Parse("https://api.giphy.com/v1/gifs/search")
	if err != nil {
		return "", err
	}

	vals := url.Values{}
	vals.Add("api_key", g.token)
	vals.Add("q", query)
	u.RawQuery = vals.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var responseData struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(jsonBytes, &responseData)
	if err != nil {
		return "", err
	}

	// Randomly select one of the available GIFs
	selectedGif := responseData.Data[rand.Intn(len(responseData.Data))]

	return selectedGif.URL, nil
}
