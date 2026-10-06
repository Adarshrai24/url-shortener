package models

type Url struct {
	ID       int    `json: id`
	Key      string `json: "key"`
	ShortURL string `json: "short_url"`
	LongURL  string `json: "long_url"`
}
