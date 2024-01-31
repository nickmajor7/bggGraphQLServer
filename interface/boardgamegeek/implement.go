package boardgamegeek

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nickmajor7/bggGraphQLServer/graph/model"
	"github.com/nickmajor7/bggGraphQLServer/interface/logger"
)

func FetchCollection(ctx context.Context, name string) (collection *model.Collection, err error) {
	var (
		items  Items
		data   []byte
		errors Errors
	)

	for {
		data, err = requestXMLAPI(ctx, name)
		if err != nil {
			if reqErr, ok := err.(ReqeustError); ok {
				if reqErr.Code == http.StatusAccepted {
					time.Sleep(1 * time.Second)
					continue
				} else {
					return
				}
			} else {
				return
			}
		} else {
			break
		}
	}

	err = xml.Unmarshal(data, &items)
	if err != nil {
		logger.Error.Println(err.Error())
		err = xml.Unmarshal(data, &errors)
		if err != nil {
			logger.Error.Println(err.Error())
			return
		} else {
			err = fmt.Errorf(errors.Error.Message)
			return
		}
	}

	return generateCollection(items, name)
}

func requestXMLAPI(ctx context.Context, name string) (data []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CollectionUrl, nil)
	if err != nil {
		err = ReqeustError{
			URL:  req.URL.String(),
			Code: http.StatusBadGateway,
			Msg:  err.Error(),
		}
		return
	}
	q := req.URL.Query()
	q.Add("username", name)
	q.Add("subtype", "boardgame")
	q.Add("stats", "1")
	q.Add("wanttoplay", "1")
	q.Add("excludesubtype", "boardgameexpansion")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		err = ReqeustError{
			URL:  req.URL.String(),
			Code: http.StatusBadGateway,
			Msg:  err.Error(),
		}
		return
	}
	if resp.StatusCode == http.StatusAccepted {
		err = ReqeustError{
			URL:  req.URL.String(),
			Code: resp.StatusCode,
			Msg:  resp.Status,
		}
		return
	} else if resp.StatusCode != http.StatusOK {
		err = ReqeustError{
			URL:  req.URL.String(),
			Code: resp.StatusCode,
			Msg:  resp.Status,
		}
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		err = ReqeustError{
			URL:  req.URL.String(),
			Code: resp.StatusCode,
			Msg:  err.Error(),
		}
	}
	return
}

func generateCollection(items Items, name string) (collection *model.Collection, err error) {
	var game *model.Game
	collection = &model.Collection{
		User: &model.User{
			Name: name,
		},
		Game: make([]*model.Game, 0),
	}
	for _, item := range items.Game {
		if game, err = generateGame(item); err != nil {
			return
		} else {
			collection.Game = append(collection.Game, game)
		}
	}

	return
}

func generateGame(item Item) (game *model.Game, err error) {
	game = &model.Game{
		ID:   item.ID,
		Name: item.Name,
	}

	game.Maxplayers, err = strconv.Atoi(item.Stats.MaxPlayers)
	if err != nil {
		return
	}

	game.Minplayers, err = strconv.Atoi(item.Stats.MinPlayers)
	if err != nil {
		return
	}

	game.Playingtime, err = strconv.Atoi(item.Stats.PlayingTime)
	if err != nil {
		return
	}

	game.Score, err = strconv.ParseFloat(item.Stats.Rate.Score.Value, 64)
	if err != nil {
		return
	}

	game.Yearpublished = fmt.Sprintf("%d", item.YearPublished)
	return
}
