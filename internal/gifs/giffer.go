package gifs

type Giffer interface {
	Gif(string) (string, error)
}
