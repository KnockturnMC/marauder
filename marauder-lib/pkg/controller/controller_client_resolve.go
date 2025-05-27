package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// ResolveArtefactReference resolves a reference to a specific artefact to its uuid.
func (h *HTTPClient) ResolveArtefactReference(ctx context.Context, reference string) (uuid.UUID, error) {
	return resolveReference(ctx, h, reference, "/artefacts/%s/%s", networkmodel.ArtefactModel{}, func(t networkmodel.ArtefactModel) uuid.UUID {
		return t.UUID
	})
}

// ResolveServerReference resolves a reference to a specific server to its uuid.
func (h *HTTPClient) ResolveServerReference(ctx context.Context, reference string) (uuid.UUID, error) {
	return resolveReference(ctx, h, reference, "/servers/%s/%s", networkmodel.ServerModel{}, func(t networkmodel.ServerModel) uuid.UUID {
		return t.UUID
	})
}

// resolveReference resolves the passed reference to either the uuid or the namespace/key layout.
func resolveReference[T any](
	ctx context.Context,
	client *HTTPClient,
	reference string,
	requestURL string,
	bindTarget T,
	mapper func(t T) uuid.UUID,
) (uuid.UUID, error) {
	uuidParsed, err := uuid.Parse(reference)
	if err == nil {
		return uuidParsed, nil
	}

	stringSplit := strings.SplitN(reference, "/", 2)
	if len(stringSplit) != 2 {
		return [16]byte{}, fmt.Errorf("failed to parse a/b format from %s: %w", reference, ErrIncorrectReferenceFormat)
	}

	result, err := utils.HTTPGetAndBind(
		ctx,
		client.Client,
		client.ControllerURL+fmt.Sprintf(requestURL, stringSplit[0], stringSplit[1]),
		bindTarget,
	)
	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to fetch and bind reference from controller: %w", err)
	}

	return mapper(result), nil
}
