package content

import "encoding/json"

type BlockType string

const (
	BlockTypeText    BlockType = "text"
	BlockTypeYouTube BlockType = "youtube"
	BlockTypeImage   BlockType = "image"
	BlockTypeVideo   BlockType = "video"
)

type Block struct {
	Type BlockType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

type TextBlockData struct {
	HTML string `json:"html"`
}

type YouTubeBlockData struct {
	YouTubeID string `json:"youtube_id"`
}

type ImageBlockData struct {
	MediaID int64  `json:"media_id"`
	Caption string `json:"caption,omitempty"`
}

type VideoBlockData struct {
	MediaID int64  `json:"media_id"`
	Caption string `json:"caption,omitempty"`
}
