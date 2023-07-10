package gif

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

type Giphy struct {
	token string
}

type GiphyResp struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func NewGiphy(token string) *Giphy {
	return &Giphy{token: token}
}

func (g *Giphy) Fetch(ctx context.Context, query string) (string, error) {
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

	jsonBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	giphyResp := &GiphyResp{}
	err = json.Unmarshal(jsonBytes, &giphyResp)
	if err != nil {
		return "", err
	}

	// Randomly select one of the available GIFs
	selectedGif := giphyResp.Data[rand.Intn(len(giphyResp.Data))]

	return selectedGif.URL, nil
}
