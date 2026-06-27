package content

import (
	"encoding/json"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
)

// SanitizeHTML trims and allowlists basic inline formatting tags.
func SanitizeHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return sanitizeAllowedHTML(raw)
}

func sanitizeAllowedHTML(raw string) string {
	allowed := []string{"p", "br", "strong", "em", "ul", "ol", "li", "a", "h2", "h3", "h4"}
	for _, tag := range allowed {
		raw = strings.ReplaceAll(raw, "<"+tag+">", "<"+tag+">")
	}
	return raw
}

// ExtractMediaIDs returns media IDs referenced by image and video blocks.
func ExtractMediaIDs(blocks []contentdomain.Block) []int64 {
	seen := make(map[int64]struct{})
	var ids []int64
	for _, block := range blocks {
		switch block.Type {
		case contentdomain.BlockTypeImage:
			var data contentdomain.ImageBlockData
			if json.Unmarshal(block.Data, &data) == nil && data.MediaID > 0 {
				if _, ok := seen[data.MediaID]; !ok {
					seen[data.MediaID] = struct{}{}
					ids = append(ids, data.MediaID)
				}
			}
		case contentdomain.BlockTypeVideo:
			var data contentdomain.VideoBlockData
			if json.Unmarshal(block.Data, &data) == nil && data.MediaID > 0 {
				if _, ok := seen[data.MediaID]; !ok {
					seen[data.MediaID] = struct{}{}
					ids = append(ids, data.MediaID)
				}
			}
		}
	}
	return ids
}

// ValidateBlocks ensures block payloads are well-formed.
func ValidateBlocks(blocks []contentdomain.Block) error {
	if len(blocks) == 0 {
		return domainerrors.ErrInvalidContentBlocks
	}

	for _, block := range blocks {
		switch block.Type {
		case contentdomain.BlockTypeText:
			var data contentdomain.TextBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil || strings.TrimSpace(data.HTML) == "" {
				return domainerrors.ErrInvalidContentBlocks
			}
		case contentdomain.BlockTypeYouTube:
			var data contentdomain.YouTubeBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil {
				return domainerrors.ErrInvalidContentBlocks
			}
			if _, err := ParseYouTubeID(data.YouTubeID); err != nil {
				return err
			}
		case contentdomain.BlockTypeImage:
			var data contentdomain.ImageBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil || data.MediaID <= 0 {
				return domainerrors.ErrInvalidContentBlocks
			}
		case contentdomain.BlockTypeVideo:
			var data contentdomain.VideoBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil || data.MediaID <= 0 {
				return domainerrors.ErrInvalidContentBlocks
			}
		default:
			return domainerrors.ErrInvalidContentBlocks
		}
	}
	return nil
}

// NormalizeBlocks sanitizes text blocks and normalizes YouTube IDs.
func NormalizeBlocks(blocks []contentdomain.Block) ([]contentdomain.Block, error) {
	if err := ValidateBlocks(blocks); err != nil {
		return nil, err
	}

	out := make([]contentdomain.Block, 0, len(blocks))
	for _, block := range blocks {
		switch block.Type {
		case contentdomain.BlockTypeText:
			var data contentdomain.TextBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil {
				return nil, domainerrors.ErrInvalidContentBlocks
			}
			data.HTML = SanitizeHTML(data.HTML)
			raw, err := json.Marshal(data)
			if err != nil {
				return nil, domainerrors.ErrInvalidContentBlocks
			}
			out = append(out, contentdomain.Block{Type: block.Type, Data: raw})
		case contentdomain.BlockTypeYouTube:
			var data contentdomain.YouTubeBlockData
			if err := json.Unmarshal(block.Data, &data); err != nil {
				return nil, domainerrors.ErrInvalidContentBlocks
			}
			id, err := ParseYouTubeID(data.YouTubeID)
			if err != nil {
				return nil, err
			}
			data.YouTubeID = id
			raw, err := json.Marshal(data)
			if err != nil {
				return nil, domainerrors.ErrInvalidContentBlocks
			}
			out = append(out, contentdomain.Block{Type: block.Type, Data: raw})
		default:
			out = append(out, block)
		}
	}
	return out, nil
}
