package app

import (
	"specfirst/internal/repository"
)

func (app *Application) CreateSnapshot(version string, tags []string, notes string) error {
	repo := repository.NewSnapshotRepository(repository.ArchivesPath())
	params := repository.CreateParams{
		Config:   app.Config,
		Protocol: app.Protocol,
		State:    app.State,
	}
	return repo.Create(version, tags, notes, params)
}

func (app *Application) RestoreSnapshot(version string, force bool) error {
	repo := repository.NewSnapshotRepository(repository.ArchivesPath())
	return repo.Restore(version, force)
}

func (app *Application) ListSnapshots() ([]string, error) {
	repo := repository.NewSnapshotRepository(repository.ArchivesPath())
	return repo.List()
}

func (app *Application) CompareSnapshots(v1, v2 string) ([]string, []string, []string, error) {
	repo := repository.NewSnapshotRepository(repository.ArchivesPath())
	return repo.Compare(v1, v2)
}
