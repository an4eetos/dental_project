package converters

import (
	"encoding/json"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	contentuc "github.com/anuarkuanysh/dental_project/internal/usecase/content"
)

type ContentBlockResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ContentItemResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Access      string                 `json:"access"`
	Locked      bool                   `json:"locked"`
	Blocks      []ContentBlockResponse `json:"blocks"`
	SortOrder   int                    `json:"sort_order,omitempty"`
}

type ContentListResponse struct {
	Items []ContentItemResponse `json:"items"`
}

func ToContentItemResponse(view contentuc.ItemView) ContentItemResponse {
	blocks := make([]ContentBlockResponse, 0, len(view.Blocks))
	for _, block := range view.Blocks {
		blocks = append(blocks, ContentBlockResponse{
			Type: string(block.Type),
			Data: block.Data,
		})
	}
	return ContentItemResponse{
		ID:          view.ID,
		Title:       view.Title,
		Description: view.Description,
		Access:      view.Access.String(),
		Locked:      view.Locked,
		Blocks:      blocks,
		SortOrder:   view.SortOrder,
	}
}

func ToContentListResponse(items []contentuc.ItemView) ContentListResponse {
	out := make([]ContentItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ToContentItemResponse(item))
	}
	return ContentListResponse{Items: out}
}

type AdminContentItemResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Access      string                 `json:"access"`
	Published   bool                   `json:"published"`
	SortOrder   int                    `json:"sort_order"`
	Blocks      []ContentBlockResponse `json:"blocks"`
}

type AdminContentListResponse struct {
	Items []AdminContentItemResponse `json:"items"`
}

func ToAdminContentItemResponse(item contentdomain.ContentItem) AdminContentItemResponse {
	blocks := make([]ContentBlockResponse, 0, len(item.Blocks))
	for _, block := range item.Blocks {
		blocks = append(blocks, ContentBlockResponse{
			Type: string(block.Type),
			Data: block.Data,
		})
	}
	return AdminContentItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Access:      item.Access.String(),
		Published:   item.Published,
		SortOrder:   item.SortOrder,
		Blocks:      blocks,
	}
}

func ToAdminContentListResponse(items []contentdomain.ContentItem) AdminContentListResponse {
	out := make([]AdminContentItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ToAdminContentItemResponse(item))
	}
	return AdminContentListResponse{Items: out}
}

type UploadMediaResponse struct {
	MediaID int64 `json:"media_id"`
}
