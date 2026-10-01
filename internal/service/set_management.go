package service

import "context"

func (t *Trainer) DeleteUnusedSet(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.repo.DeleteUnusedSet(ctx, id); err != nil {
		return err
	}
	// A previously prepared duplicate import must not silently recreate a deleted set.
	t.preview = nil
	t.token = ""
	return nil
}
func (t *Trainer) SetArchived(ctx context.Context, id string, archived bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.SetArchived(ctx, id, archived)
}
