package content

// AccessLevel controls who can view full content.
type AccessLevel string

const (
	AccessPublic       AccessLevel = "public"
	AccessSubscription AccessLevel = "subscription"
)

func (a AccessLevel) String() string {
	return string(a)
}

func ParseAccessLevel(raw string) (AccessLevel, bool) {
	switch AccessLevel(raw) {
	case AccessPublic, AccessSubscription:
		return AccessLevel(raw), true
	default:
		return "", false
	}
}

// IsAccessible reports whether the user may view full content blocks.
func IsAccessible(access AccessLevel, subscriptionActive bool) bool {
	if access == AccessPublic {
		return true
	}
	return subscriptionActive
}
