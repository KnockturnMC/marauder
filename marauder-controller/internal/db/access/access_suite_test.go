package access_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/avast/retry-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Include postgres driver.
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
)

const MARAUDER = "marauder"

var (
	databaseClient       *sqlm.DB
	dockerContainer      container.CreateResponse
	dockerClientInstance *dockerClient.Client
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

	dockerClientInstance, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	Expect(err).To(Not(HaveOccurred()))

	port, err := freeport.GetFreePort()
	Expect(err).To(Not(HaveOccurred()))

	databaseStartupCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancelFunc()

	dockerContainer, err = dockerClientInstance.ContainerCreate(
		databaseStartupCtx,
		&container.Config{
			Image:        "postgres",
			ExposedPorts: map[nat.Port]struct{}{"5432": {}},
			Env: []string{
				"POSTGRES_PASSWORD=" + MARAUDER,
				"POSTGRES_USER=" + MARAUDER,
			},
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{"5432/tcp": {nat.PortBinding{HostPort: strconv.Itoa(port), HostIP: "0.0.0.0"}}},
			AutoRemove:   true,
		},
		nil,
		nil,
		fmt.Sprintf("marauder-postgres-docker-%d", port),
	)
	Expect(err).To(Not(HaveOccurred()))

	err = dockerClientInstance.ContainerStart(databaseStartupCtx, dockerContainer.ID, types.ContainerStartOptions{})
	Expect(err).To(Not(HaveOccurred()))

	databaseConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"0.0.0.0", port, MARAUDER, MARAUDER, MARAUDER,
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
