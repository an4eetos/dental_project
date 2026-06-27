package content

import (
	"context"
	"encoding/json"
	"testing"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type stubContentRepo struct {
	items []contentdomain.ContentItem
}

func (s *stubContentRepo) ListPublished(context.Context) ([]contentdomain.ContentItem, error) {
	return s.items, nil
}

func (s *stubContentRepo) ListAll(context.Context) ([]contentdomain.ContentItem, error) {
	return s.items, nil
}

func (s *stubContentRepo) GetByID(context.Context, int64) (contentdomain.ContentItem, error) {
	return contentdomain.ContentItem{}, nil
}

func (s *stubContentRepo) Create(context.Context, port.CreateContentParams) (contentdomain.ContentItem, error) {
	return contentdomain.ContentItem{}, nil
}

func (s *stubContentRepo) Update(context.Context, int64, port.UpdateContentParams) (contentdomain.ContentItem, error) {
	return contentdomain.ContentItem{}, nil
}

func (s *stubContentRepo) Delete(context.Context, int64) error { return nil }

func (s *stubContentRepo) UpdateSortOrder(context.Context, []int64) error { return nil }

func (s *stubContentRepo) NextSortOrder(context.Context) (int, error) { return 1, nil }

type stubUsers struct {
	user identity.User
}

func (s *stubUsers) UpsertByTelegramID(context.Context, port.UpsertUserParams) (identity.User, error) {
	return identity.User{}, nil
}

func (s *stubUsers) GetByID(context.Context, int64) (identity.User, error) {
	return s.user, nil
}

func (s *stubUsers) GetByTelegramID(context.Context, int64) (identity.User, error) {
	return identity.User{}, nil
}

func (s *stubUsers) List(context.Context, port.ListUsersParams) ([]identity.User, error) {
	return nil, nil
}

func (s *stubUsers) GetOverviewByID(context.Context, int64) (admin.UserOverview, error) {
	return admin.UserOverview{}, nil
}

func (s *stubUsers) UpdateByAdmin(context.Context, int64, port.AdminUpdateUserParams) (identity.User, error) {
	return identity.User{}, nil
}

func (s *stubUsers) SetBlocked(context.Context, int64, bool) (identity.User, error) {
	return identity.User{}, nil
}

type stubChecker struct {
	active bool
}

func (s *stubChecker) Check(context.Context, identity.User) (port.SubscriptionStatus, error) {
	return port.SubscriptionStatus{Active: s.active}, nil
}

func TestListPublished_MasksSubscriptionContent(t *testing.T) {
	t.Parallel()

	item := contentdomain.ContentItem{
		ID:     1,
		Title:  "Premium",
		Access: contentdomain.AccessSubscription,
		Blocks: []contentdomain.Block{
			{Type: contentdomain.BlockTypeYouTube, Data: json.RawMessage(`{"youtube_id":"zQZ3SGSwGBI"}`)},
		},
	}

	uc := &ListPublished{
		Content: &stubContentRepo{items: []contentdomain.ContentItem{item}},
		Checker: &stubChecker{active: false},
		Users:   &stubUsers{user: identity.User{ID: 1, Role: identity.RolePatient}},
	}

	views, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 item, got %d", len(views))
	}
	if !views[0].Locked {
		t.Fatal("expected locked view")
	}

	var data contentdomain.YouTubeBlockData
	if err := json.Unmarshal(views[0].Blocks[0].Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.YouTubeID != "" {
		t.Fatalf("expected masked youtube id, got %q", data.YouTubeID)
	}
}
