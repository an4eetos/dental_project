package content

import (
	"encoding/json"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
)

// MaskBlocks redacts premium block payloads for users without entitlement.
func MaskBlocks(blocks []contentdomain.Block) []contentdomain.Block {
	out := make([]contentdomain.Block, 0, len(blocks))
	for _, block := range blocks {
		switch block.Type {
		case contentdomain.BlockTypeYouTube:
			out = append(out, contentdomain.Block{
				Type: block.Type,
				Data: json.RawMessage(`{"youtube_id":""}`),
			})
		case contentdomain.BlockTypeImage:
			out = append(out, contentdomain.Block{
				Type: block.Type,
				Data: json.RawMessage(`{"media_id":0}`),
			})
		case contentdomain.BlockTypeVideo:
			out = append(out, contentdomain.Block{
				Type: block.Type,
				Data: json.RawMessage(`{"media_id":0}`),
			})
		default:
			out = append(out, block)
		}
	}
	return out
}
