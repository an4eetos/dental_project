package content

import "time"

// ContentItem is an editable educational material shown in the patient app.
type ContentItem struct {
	ID          int64
	Title       string
	Description string
	Access      AccessLevel
	Published   bool
	SortOrder   int
	Blocks      []Block
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
