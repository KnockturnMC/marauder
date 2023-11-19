package access

import (
	"context"
	"fmt"

	"github.com/Goldziher/go-utils/sliceutils"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/google/uuid"
)

// FindServerTargetStateMissMatch fetches a miss-match between the servers current is states and target state.
func FindServerTargetStateMissMatch(
	ctx context.Context,
	db *sqlm.DB,
	server uuid.UUID,
	requiresRestart bool,
) ([]networkmodel.ArtefactVersionMissmatch, error) {
	type DBArtefactVersionMissmatchModel struct {
		ArtefactIdentifier string     `db:"artefact_identifier"`
		IsArtefact         *uuid.UUID `db:"is_artefact"`
		IsVersion          *string    `db:"is_version"`
		TargetArtefact     *uuid.UUID `db:"target_artefact"`
		TargetVersion      *string    `db:"target_version"`
		RequiresRestart    bool       `db:"requires_restart"`
	}

	result := make([]DBArtefactVersionMissmatchModel, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM func_find_server_target_state_missmatches($1) WHERE ($2 OR NOT requires_restart)
		`, server, requiresRestart); err != nil {
		return nil, fmt.Errorf("failed to execute psql func to fetch missmatch: %w", err)
	}

	return sliceutils.Map(
		result,
		func(
			value DBArtefactVersionMissmatchModel,
			index int,
			slice []DBArtefactVersionMissmatchModel,
		) networkmodel.ArtefactVersionMissmatch {
			var isArtefact, targetArtefact *networkmodel.ArtefactVersionMissmatchArtefactInfo
			var missmatch networkmodel.ArtefactMissmatch
			if value.TargetArtefact != nil {
				targetArtefact = &networkmodel.ArtefactVersionMissmatchArtefactInfo{
					Artefact: *value.TargetArtefact,
					Version:  *value.TargetVersion,
				}
			}
			if value.IsArtefact != nil {
				isArtefact = &networkmodel.ArtefactVersionMissmatchArtefactInfo{
					Artefact: *value.IsArtefact,
					Version:  *value.IsVersion,
				}
			}

			switch {
			case isArtefact != nil && targetArtefact != nil:
				missmatch = networkmodel.ArtefactMissmatch{
					Update: &networkmodel.ArtefactVersionMissmatchUpdate{
						Is: *isArtefact, Target: *targetArtefact,
					},
				}
			case isArtefact != nil:
				missmatch = networkmodel.ArtefactMissmatch{
					Uninstall: &networkmodel.ArtefactVersionMissmatchUninstall{Is: *isArtefact},
				}
			default:
				missmatch = networkmodel.ArtefactMissmatch{
					Install: &networkmodel.ArtefactVersionMissmatchInstall{Target: *targetArtefact},
				}
			}

			return networkmodel.ArtefactVersionMissmatch{
				ArtefactIdentifier: value.ArtefactIdentifier,
				Missmatch:          missmatch,
			}
		},
	), nil
}

// FetchServerArtefactsByState fetches all artefacts for a given server and the given state.
func FetchServerArtefactsByState(
	ctx context.Context,
	db *sqlm.DB,
	server uuid.UUID,
	state networkmodel.ServerStateType,
) ([]networkmodel.ArtefactModel, error) {
	result := make([]networkmodel.ArtefactModel, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM func_find_server_artefacts_by_state($1, $2)
		`, server, state); err != nil {
		return nil, fmt.Errorf("failed to fetch server artefacts by state: %w", err)
	}

	return result, nil
}

// UpdateDeployment creates a state in the state table for the given server to the new artefact
// and ensures that the existing indices are kept in place.
func UpdateDeployment(
	ctx context.Context,
	db *sqlm.DB,
	model networkmodel.ServerArtefactStateModel,
) (networkmodel.ServerArtefactStateModel, error) {
	if !networkmodel.KnownServerStateType(model.Type) {
		return networkmodel.ServerArtefactStateModel{}, fmt.Errorf("unknown server state (%s): %w", model.Type, networkmodel.ErrUnknownServerState)
	}

	if err := db.NamedGetContext(ctx, &model, `
		SELECT * FROM func_create_server_state(:server, :artefact_identifier, :artefact_uuid, :type) 
		`, model); err != nil {
		return networkmodel.ServerArtefactStateModel{}, fmt.Errorf("failed to create new server state: %w", err)
	}

	return model, nil
}
