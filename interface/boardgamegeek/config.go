package boardgamegeek

import (
	"net/http"
	"time"
)

const (
	CollectionUrl = "https://boardgamegeek.com/xmlapi2/collection"
)

var (
	client = http.Client{
		Timeout: 30 * time.Second,
	}
)
