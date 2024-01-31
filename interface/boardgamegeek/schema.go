package boardgamegeek

import (
	"encoding/xml"
	"fmt"
)

type Items struct {
	XMLName  xml.Name `xml:"items"`
	TotalNum string   `xml:"totalitems,attr"`
	Game     []Item   `xml:"item"`
}

type Errors struct {
	XMLName xml.Name `xml:"errors"`
	Error   Error    `xml:"error"`
}

type Error struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message"`
}

type Item struct {
	ID            string   `xml:"objectid,attr"`
	XMLName       xml.Name `xml:"item"`
	Name          string   `xml:"name"`
	YearPublished int      `xml:"yearpublished"`
	Stats         Stats    `xml:"stats"`
}

type Stats struct {
	XMLName     xml.Name `xml:"stats"`
	MinPlayers  string   `xml:"minplayers,attr"`
	MaxPlayers  string   `xml:"maxplayers,attr"`
	PlayingTime string   `xml:"playingtime,attr"`
	Rate        Rateing  `xml:"rating"`
}

type Rateing struct {
	XMLName xml.Name     `xml:"rating"`
	Score   Bayesaverage `xml:"bayesaverage"`
}

type Bayesaverage struct {
	XMLName xml.Name `xml:"bayesaverage"`
	Value   string   `xml:"value,attr"`
}

type ReqeustError struct {
	URL  string
	Code int
	Msg  string
}

// STEP 2：定義能夠屬於 Error Interface 的方法
func (e ReqeustError) Error() string {
	return fmt.Sprintf("request %s code %d", e.URL, e.Code)
}
