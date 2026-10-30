package story

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	items []Progress
	saved *Progress
}

func (f *fakeStore) ListByUser(context.Context, string) ([]Progress, error) {
	return f.items, nil
}

func (f *fakeStore) Get(context.Context, string, string) (*Progress, error) {
	return f.saved, nil
}

func (f *fakeStore) Upsert(_ context.Context, progress *Progress) error {
	f.saved = progress
	return nil
}

func TestSaveDefaultsFlagsToEmptyObject(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store)
	fixedNow := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	progress, err := service.Save(context.Background(), "user-1", SaveInput{
		ChapterID: "chapter-1",
		NodeID:    "node-1",
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if string(progress.Flags) != "{}" {
		t.Fatalf("flags = %s, want {}", progress.Flags)
	}
	if !progress.UpdatedAt.Equal(fixedNow) {
		t.Fatalf("updated_at = %s, want %s", progress.UpdatedAt, fixedNow)
	}
}

func TestSaveRejectsNonObjectFlags(t *testing.T) {
	service := NewService(&fakeStore{})

	_, err := service.Save(context.Background(), "user-1", SaveInput{
		ChapterID: "chapter-1",
		NodeID:    "node-1",
		Flags:     []byte(`[1,2]`),
	})
	if !errors.Is(err, ErrInvalidFlags) {
		t.Fatalf("err = %v, want ErrInvalidFlags", err)
	}
}

func TestSaveRejectsOversizedFlags(t *testing.T) {
	service := NewService(&fakeStore{})
	flags := []byte(`{"pad":"` + strings.Repeat("x", maxFlagsBytes) + `"}`)

	_, err := service.Save(context.Background(), "user-1", SaveInput{
		ChapterID: "chapter-1",
		NodeID:    "node-1",
		Flags:     flags,
	})
	if !errors.Is(err, ErrInvalidFlags) {
		t.Fatalf("err = %v, want ErrInvalidFlags", err)
	}
}

func TestSaveRequiresChapterAndNode(t *testing.T) {
	service := NewService(&fakeStore{})

	_, err := service.Save(context.Background(), "user-1", SaveInput{ChapterID: "chapter-1"})
	if !errors.Is(err, ErrInvalidProgress) {
		t.Fatalf("err = %v, want ErrInvalidProgress", err)
	}
}
