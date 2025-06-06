CREATE DOMAIN CREATION_DATE AS TIMESTAMPTZ DEFAULT NOW()::TIMESTAMPTZ(0);
CREATE DOMAIN CPU_CORE AS INTEGER CHECK ( value >= 0);
CREATE DOMAIN SHA_256_HASH AS BYTEA CHECK ( length(value) = 32 );
CREATE TYPE SERVER_STATE_TYPE AS ENUM ('TARGET', 'IS', 'HISTORY');

CREATE TABLE artefact
(
	uuid             UUID          NOT NULL DEFAULT gen_random_uuid(),
	identifier       VARCHAR       NOT NULL,
	version          VARCHAR       NOT NULL,
	upload_date      CREATION_DATE NOT NULL,
	requires_restart BOOLEAN       NOT NULL,

	CONSTRAINT pk_artefact PRIMARY KEY (uuid),
	CONSTRAINT un_artefact_identifier_version UNIQUE (identifier, version)
);

CREATE TABLE artefact_file
(
	artefact UUID         NOT NULL,
	tarball  BYTEA        NOT NULL,
	hash     SHA_256_HASH NOT NULL,

	CONSTRAINT pk_artefact_file PRIMARY KEY (artefact),
	CONSTRAINT fk_artefact_file_artefact_uuid FOREIGN KEY (artefact) REFERENCES artefact (uuid)
		ON DELETE CASCADE
);

CREATE TABLE server_operator
(
	identifier VARCHAR NOT NULL,
	host       VARCHAR NOT NULL,
	port       INT     NOT NULL,

	CONSTRAINT pk_server_operator PRIMARY KEY (identifier),
	CONSTRAINT un_server_operator UNIQUE (host, port)
);

CREATE TABLE server
(
	uuid        UUID             NOT NULL DEFAULT gen_random_uuid(),
	environment VARCHAR          NOT NULL,
	name        VARCHAR          NOT NULL,
	operator    VARCHAR          NOT NULL,
	memory      INTEGER          NOT NULL,
	cpu         DOUBLE PRECISION NOT NULL,
	port        INTEGER          NOT NULL,
	image       VARCHAR          NOT NULL,

	CONSTRAINT pk_servers_uuid PRIMARY KEY (uuid),
	CONSTRAINT un_servers_environment_name UNIQUE (environment, name),
	CONSTRAINT fk_server_operator FOREIGN KEY (operator) REFERENCES server_operator (identifier) ON DELETE CASCADE
		ON UPDATE CASCADE
);

CREATE TABLE server_network
(
	uuid         UUID    NOT NULL DEFAULT gen_random_uuid(),
	server       UUID    NOT NULL,
	network_name VARCHAR NOT NULL,
	ipv4_address VARCHAR NOT NULL,

	CONSTRAINT pk_server_network_uuid PRIMARY KEY (uuid),
	CONSTRAINT un_server_network_server_network_name UNIQUE (server, network_name),
	CONSTRAINT fk_server_network_server FOREIGN KEY (server) REFERENCES server (uuid) ON DELETE CASCADE
		ON UPDATE CASCADE
);

CREATE TABLE server_host_port
(
	uuid        UUID    NOT NULL DEFAULT gen_random_uuid(),
	server      UUID    NOT NULL,
	host_ip     VARCHAR NOT NULL,
	host_port   INT     NOT NULL,
	server_port INT     NOT NULL,

	CONSTRAINT pk_server_host_port_uuid PRIMARY KEY (uuid),
	CONSTRAINT un_server_host_port_host_ip_port UNIQUE (host_ip, host_port),
	CONSTRAINT fk_server_host_port_server FOREIGN KEY (server) REFERENCES server (uuid) ON DELETE CASCADE
		ON UPDATE CASCADE,
	CONSTRAINT un_server_host_port_server_port UNIQUE (server, server_port)
);

CREATE TABLE server_state
(
	uuid                UUID              NOT NULL DEFAULT gen_random_uuid(),
	server              UUID              NOT NULL,
	artefact_identifier VARCHAR           NOT NULL,
	artefact_uuid       UUID              NOT NULL,
	definition_date     CREATION_DATE     NOT NULL,
	type                SERVER_STATE_TYPE NOT NULL,

	CONSTRAINT pk_server_state PRIMARY KEY (uuid),
	CONSTRAINT fk_server FOREIGN KEY (server) REFERENCES server (uuid)
		ON DELETE CASCADE,
	CONSTRAINT fk_artefact_uuid FOREIGN KEY (artefact_uuid) REFERENCES artefact (uuid)
		ON DELETE CASCADE
);

CREATE VIEW server_state_target AS
SELECT *
FROM server_state
WHERE type = 'TARGET';

CREATE VIEW server_state_is AS
SELECT *
FROM server_state
WHERE type = 'IS';

CREATE INDEX idx_server_state_non_history_state_server ON server_state (server) WHERE type != 'HISTORY';
CREATE INDEX idx_server_state_non_history_state_artefact_uuid ON server_state (artefact_uuid) WHERE type != 'HISTORY';
CREATE INDEX idx_server_state_non_history_state_artefact_identifier ON server_state (artefact_identifier) WHERE type != 'HISTORY';
CREATE UNIQUE INDEX idx_server_state_is_uniq ON server_state (server, artefact_identifier) WHERE type = 'IS';
CREATE UNIQUE INDEX idx_server_state_target_uniq ON server_state (server, artefact_identifier) WHERE type = 'TARGET';

CREATE TABLE cronjob
(
	type           VARCHAR   NOT NULL,
	next_execution TIMESTAMP NOT NULL,

	CONSTRAINT pk_cronjob PRIMARY KEY (type)
)
