package access_test

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	_ "github.com/lib/pq" // Include postgres driver.
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
)

const (
	Marauder      = "marauder"
	DBDockerImage = "postgres:15.3"
)

var (
	databaseClient       *sqlm.DB
	dockerContainer      container.CreateResponse
	dockerClientInstance *client.Client
	ctx                  context.Context
)

func TestAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Parallel()
	RunSpecs(t, "Access Suite")
}

var _ = BeforeSuite(func() {
	if !Label("functiontest").MatchesLabelFilter(GinkgoLabelFilter()) {
		return
	}

	ctx = context.Background()
	var err error

	dockerClientInstance, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	Expect(err).To(Not(HaveOccurred()))

	port, err := freeport.GetFreePort()
	Expect(err).To(Not(HaveOccurred()))

	databaseStartupCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(2*time.Minute))
	defer cancelFunc()

	pullReader, err := dockerClientInstance.ImagePull(databaseStartupCtx, DBDockerImage, image.PullOptions{})
	Expect(err).To(Not(HaveOccurred()))
	_, _ = io.ReadAll(pullReader)
	defer func() { _ = pullReader.Close() }()

	dockerContainer, err = dockerClientInstance.ContainerCreate(
		databaseStartupCtx,
		&container.Config{
			Image:        DBDockerImage,
			ExposedPorts: map[nat.Port]struct{}{"5432": {}},
			Env: []string{
				"POSTGRES_PASSWORD=" + Marauder,
				"POSTGRES_USER=" + Marauder,
				"POSTGRES_DB=" + Marauder,
			},
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{"5432/tcp": {nat.PortBinding{HostPort: strconv.Itoa(port), HostIP: "0.0.0.0"}}},
			AutoRemove:   true,
		},
		nil,
		nil,
		fmt.Sprintf("marauder-postgres-docker-test-%d", port),
	)
	Expect(err).To(Not(HaveOccurred()))

	err = dockerClientInstance.ContainerStart(databaseStartupCtx, dockerContainer.ID, container.StartOptions{})
	Expect(err).To(Not(HaveOccurred()))

	databaseConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable binary_parameters=yes",
		"0.0.0.0", port, Marauder, Marauder, Marauder,
	)

	err = retry.Do(func() error {
		connection, err := sqlx.ConnectContext(databaseStartupCtx, "postgres", databaseConnectionString)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		databaseClient = &sqlm.DB{DB: connection}

		return nil
	}, retry.Context(databaseStartupCtx))
	Expect(err).To(Not(HaveOccurred()))

	instance, err := postgres.WithInstance(databaseClient.DB.DB, &postgres.Config{})
	Expect(err).To(Not(HaveOccurred()))

	Expect(sqlm.ApplyMigrations(instance, "marauder")).To(Not(HaveOccurred()))
})

var _ = AfterSuite(func() {
	if !Label("functiontest").MatchesLabelFilter(GinkgoLabelFilter()) {
		return
	}

	ctx := context.Background()
	err := dockerClientInstance.ContainerKill(ctx, dockerContainer.ID, "SIGKILL")
	Expect(err).To(Not(HaveOccurred()))
})
