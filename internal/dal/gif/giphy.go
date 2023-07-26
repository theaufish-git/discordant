package gif

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Giphy struct {
	service.UnimplementedService
	token string
}

type GiphyResp struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func NewGiphy(cfg *config.Gif) *Giphy {
	return &Giphy{token: cfg.Token}
}

func (g *Giphy) Initialize(ctx context.Context) error {
	return nil
}

func (g *Giphy) Shutdown(ctx context.Context) error {
	return nil
}

func (g *Giphy) String() string {
	return "giphy"
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
