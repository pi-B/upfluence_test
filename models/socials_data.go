package models

var SOCIAL_TYPES = []string{"instagram_media", "pin", "youtube_video", "article", "tweet", "facebook_status"}
var DIMENSION_TYPE = []string{"likes", "comments", "favorites", "retweets"}

type SocialsData struct {
	Id        int `json:"id"`
	Likes     int `json:"likes,omitempty"`
	Comments  int `json:"comments,omitempty"`
	Favorites int `json:"favorites,omitempty"`
	Retweet   int `json:"retweets,omitempty"`
	Timestamp int `json:"timestamp"`
}
