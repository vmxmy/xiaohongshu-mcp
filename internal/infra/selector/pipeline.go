package selector

import "context"

type Learner interface {
	Learn(ctx context.Context, current map[string]string) (map[string]string, error)
}

type Validator interface {
	Validate(ctx context.Context, selectors map[string]string) error
}

type Store interface {
	Load() (map[string]string, error)
	Save(selectors map[string]string) error
	Snapshot() (string, error)
	Rollback(snapshot string) error
}

type Pipeline struct {
	Store    Store
	Learner  Learner
	Validate Validator
}

func (p Pipeline) Run(ctx context.Context) error {
	current, err := p.Store.Load()
	if err != nil {
		return err
	}
	snap, err := p.Store.Snapshot()
	if err != nil {
		return err
	}
	next, err := p.Learner.Learn(ctx, current)
	if err != nil {
		return err
	}
	if err := p.Validate.Validate(ctx, next); err != nil {
		_ = p.Store.Rollback(snap)
		return err
	}
	return p.Store.Save(next)
}
