package photoreview

// MediaType distinguishes photo and video patient submissions.
type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
)

func (m MediaType) Valid() bool {
	switch m {
	case MediaTypePhoto, MediaTypeVideo:
		return true
	default:
		return false
	}
}

func (m MediaType) String() string { return string(m) }
