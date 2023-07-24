package servermgr

import (
	"context"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

func (d DockerBasedManager) Stop(_ context.Context, _ networkmodel.ServerModel) error {
	panic("implement me")
}
