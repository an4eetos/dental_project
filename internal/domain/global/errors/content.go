package errors

var (
	ErrContentNotFound       = BaseError{Msg: "content not found", ErrC: "content_not_found"}
	ErrInvalidContentBlocks  = BaseError{Msg: "invalid content blocks", ErrC: "invalid_content_blocks"}
	ErrInvalidYouTubeID      = BaseError{Msg: "invalid youtube id", ErrC: "invalid_youtube_id"}
	ErrContentMediaNotFound  = BaseError{Msg: "content media not found", ErrC: "content_media_not_found"}
	ErrContentMediaTooLarge  = BaseError{Msg: "content media too large", ErrC: "content_media_too_large"}
	ErrInvalidContentAccess  = BaseError{Msg: "invalid content access level", ErrC: "invalid_content_access"}
	ErrInvalidContentTitle   = BaseError{Msg: "content title is required", ErrC: "invalid_content_title"}
	ErrInvalidContentMedia   = BaseError{Msg: "invalid content media type", ErrC: "invalid_content_media"}
)
