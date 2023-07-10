package gif

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

type Tenor struct {
	token string
}

type TenorResponse struct {
	Results []struct {
		Media struct {
			Gif struct {
				URL string `json:"url"`
			} `json:"gif"`
		} `json:"media_formats"`
	} `json:"results"`
}

func NewTenor(token string) *Tenor {
	return &Tenor{token: token}
}

func (t *Tenor) Fetch(ctx context.Context, query string) (string, error) {
	u, err := url.Parse("https://tenor.googleapis.com/v2/search")
	if err != nil {
		return "", err
	}

	vals := url.Values{}
	vals.Add("q", query)
	vals.Add("key", t.token)
	vals.Add("client_key", "discordant")
	vals.Add("limit", "20")
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

	tenorResp := &TenorResponse{}
	err = json.Unmarshal(jsonBytes, tenorResp)
	if err != nil {
		return "", err
	}

	// Randomly select one of the available GIFs
	selectedGif := tenorResp.Results[rand.Intn(len(tenorResp.Results))].Media.Gif

	return selectedGif.URL, nil
}
