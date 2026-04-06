package cli

import (
	"database/sql"
	"fmt"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/sherlook22/cortex/internal/infrastructure/storage/sqlite"
)

// deps holds all use cases, wired to a single database connection.
type deps struct {
	db      *sql.DB
	save    *application.SaveMemoryUseCase
	search  *application.SearchMemoryUseCase
	get     *application.GetMemoryUseCase
	update  *application.UpdateMemoryUseCase
	del     *application.DeleteMemoryUseCase
	context *application.GetContextUseCase
	stats   *application.GetStatsUseCase
}

// newDeps opens the database and wires all dependencies.
func newDeps() (*deps, error) {
	cfg := sqlite.DefaultConfig()
	db, err := sqlite.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	repo := sqlite.NewRepository(db)

	return &deps{
		db:      db,
		save:    application.NewSaveMemoryUseCase(repo),
		search:  application.NewSearchMemoryUseCase(repo),
		get:     application.NewGetMemoryUseCase(repo),
		update:  application.NewUpdateMemoryUseCase(repo),
		del:     application.NewDeleteMemoryUseCase(repo),
		context: application.NewGetContextUseCase(repo),
		stats:   application.NewGetStatsUseCase(repo),
	}, nil
}

func (d *deps) close() {
	if d.db != nil {
		d.db.Close()
	}
}
