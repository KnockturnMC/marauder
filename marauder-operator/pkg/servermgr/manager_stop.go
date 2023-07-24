package servermgr

import (
	"context"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

func (d DockerBasedManager) Stop(ctx context.Context, server networkmodel.ServerModel) error {
	panic("implement me")
}
