CREATE DOMAIN CREATION_DATE AS TIMESTAMPTZ DEFAULT NOW()::TIMESTAMPTZ(0);
CREATE DOMAIN CPU_CORE AS INTEGER CHECK ( value >= 0);
CREATE DOMAIN SHA_256_HASH AS BYTEA CHECK ( length(value) = 32 );
CREATE TYPE SERVER_STATE_TYPE AS ENUM ('TARGET', 'IS', 'HISTORY');

CREATE TABLE artefact
(
    uuid        UUID          NOT NULL DEFAULT gen_random_uuid(),
    identifier  VARCHAR       NOT NULL,
    version     VARCHAR       NOT NULL,
    upload_date CREATION_DATE NOT NULL,

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

CREATE TABLE server
(
    uuid        UUID             NOT NULL DEFAULT gen_random_uuid(),
    environment VARCHAR          NOT NULL,
    name        VARCHAR          NOT NULL,
    host        VARCHAR          NOT NULL,
    memory      INTEGER          NOT NULL,
    cpu         DOUBLE PRECISION NOT NULL,
    image       VARCHAR          NOT NULL,

    CONSTRAINT pk_servers_uuid PRIMARY KEY (uuid),
    CONSTRAINT un_servers_environment_name UNIQUE (environment, name),
    CONSTRAINT un_server_uuid_host UNIQUE (uuid, host)
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

CREATE TABLE server_state
(
    uuid            UUID              NOT NULL DEFAULT gen_random_uuid(),
    server          UUID              NOT NULL,
    artefact        UUID              NOT NULL,
    definition_date CREATION_DATE     NOT NULL,
    type            SERVER_STATE_TYPE NOT NULL,

    CONSTRAINT pk_server_state PRIMARY KEY (uuid),
    CONSTRAINT fk_server FOREIGN KEY (server) REFERENCES server (uuid)
        ON DELETE CASCADE,
    CONSTRAINT fk_artefact FOREIGN KEY (artefact) REFERENCES artefact (uuid)
        ON DELETE CASCADE
);

CREATE INDEX idx_server_state_non_history_state_server ON server_state (server) WHERE type != 'HISTORY';
CREATE INDEX idx_server_state_non_history_state_artefact ON server_state (artefact) WHERE type != 'HISTORY';
CREATE UNIQUE INDEX idx_server_state_is_uniq ON server_state (server) WHERE type = 'IS';
CREATE UNIQUE INDEX idx_server_state_target_uniq ON server_state (server) WHERE type = 'TARGET';
