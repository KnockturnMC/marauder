package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/networkmodel"
	"github.com/google/uuid"
)

// FindServerTargetStateMissMatch fetches a miss-match between the servers current is states and target state.
func FindServerTargetStateMissMatch(ctx context.Context, db *sqlm.DB, server uuid.UUID) ([]networkmodel.VersionDiff, error) {
	result := make([]networkmodel.VersionDiff, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM func_find_server_target_state_missmatches($1)
		`, server); err != nil {
		return nil, fmt.Errorf("failed to execute psql func to fetch missmatch: %w", err)
	}

	return result, nil
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

func UpdateIsStateOfServer(
	ctx context.Context,
	db *sqlm.DB,
	server uuid.UUID,
	newIsArtefact uuid.UUID,
) error {

}
