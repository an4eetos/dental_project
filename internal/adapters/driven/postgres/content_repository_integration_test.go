//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	postgresadapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/postgres"
	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
	infrapostgres "github.com/anuarkuanysh/dental_project/infra/postgresql"
)

func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("dental_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := infrapostgres.RunMigrations(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func TestContentRepository_CRUDAndReorder(t *testing.T) {
	pool := startPostgres(t)
	repo := postgresadapter.NewContentRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, port.CreateContentParams{
		Title:       "Test item",
		Description: "Desc",
		Access:      contentdomain.AccessPublic,
		Published:   true,
		SortOrder:   1,
		Blocks: []contentdomain.Block{
			{Type: contentdomain.BlockTypeText, Data: []byte(`{"html":"<p>hello</p>"}`)},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	second, err := repo.Create(ctx, port.CreateContentParams{
		Title:     "Second",
		Access:    contentdomain.AccessSubscription,
		Published: false,
		SortOrder: 2,
		Blocks: []contentdomain.Block{
			{Type: contentdomain.BlockTypeYouTube, Data: []byte(`{"youtube_id":"zQZ3SGSwGBI"}`)},
		},
	})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	published, err := repo.ListPublished(ctx)
	if err != nil {
		t.Fatalf("list published: %v", err)
	}
	if len(published) != 1 || published[0].ID != created.ID {
		t.Fatalf("unexpected published list: %+v", published)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Test item" {
		t.Fatalf("got title %q", got.Title)
	}

	updated, err := repo.Update(ctx, created.ID, port.UpdateContentParams{
		Title:     "Updated",
		Access:    contentdomain.AccessPublic,
		Published: true,
		Blocks:    got.Blocks,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected updated title")
	}
	if updated.UpdatedAt.Before(updated.CreatedAt.Add(-time.Second)) {
		t.Fatalf("updated_at not set")
	}

	if err := repo.UpdateSortOrder(ctx, []int64{second.ID, created.ID}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 || all[0].ID != second.ID {
		t.Fatalf("unexpected order: %+v", all)
	}

	if err := repo.Delete(ctx, second.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
