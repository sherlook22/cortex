package cli

import (
	"database/sql"
	"fmt"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/sherlook22/cortex/internal/infrastructure/storage/sqlite"
)

// deps holds all use cases, wired to a single database connection.
type deps struct {
	db           *sql.DB
	save         *application.SaveMemoryUseCase
	search       *application.SearchMemoryUseCase
	get          *application.GetMemoryUseCase
	update       *application.UpdateMemoryUseCase
	del          *application.DeleteMemoryUseCase
	context      *application.GetContextUseCase
	stats        *application.GetStatsUseCase
	export       *application.ExportUseCase
	imp          *application.ImportUseCase
	capture      *application.CaptureUseCase
	sessionStart *application.StartSessionUseCase
	sessionEnd   *application.EndSessionUseCase
	sessionList  *application.ListSessionsUseCase
	sessionGet   *application.GetSessionUseCase
}

// newDeps opens the database and wires all dependencies.
func newDeps() (*deps, error) {
	cfg := sqlite.DefaultConfig()
	db, err := sqlite.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	memRepo := sqlite.NewRepository(db)
	sessRepo := sqlite.NewSessionRepository(db)

	return &deps{
		db:           db,
		save:         application.NewSaveMemoryUseCase(memRepo),
		search:       application.NewSearchMemoryUseCase(memRepo),
		get:          application.NewGetMemoryUseCase(memRepo),
		update:       application.NewUpdateMemoryUseCase(memRepo),
		del:          application.NewDeleteMemoryUseCase(memRepo),
		context:      application.NewGetContextUseCase(memRepo),
		stats:        application.NewGetStatsUseCase(memRepo),
		export:       application.NewExportUseCase(memRepo),
		imp:          application.NewImportUseCase(memRepo),
		capture:      application.NewCaptureUseCase(memRepo),
		sessionStart: application.NewStartSessionUseCase(sessRepo),
		sessionEnd:   application.NewEndSessionUseCase(sessRepo),
		sessionList:  application.NewListSessionsUseCase(sessRepo),
		sessionGet:   application.NewGetSessionUseCase(sessRepo),
	}, nil
}

func (d *deps) close() {
	if d.db != nil {
		d.db.Close()
	}
}
