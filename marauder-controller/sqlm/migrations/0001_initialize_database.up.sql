CREATE DOMAIN CREATION_DATE AS TIMESTAMPTZ DEFAULT NOW()::TIMESTAMPTZ(0);
CREATE DOMAIN CPU_CORE AS INTEGER CHECK ( value >= 0);
CREATE TYPE SERVER_STATE_TYPE AS ENUM ('TARGET', 'IS', 'HISTORY');

CREATE TABLE artefact
(
    uuid         UUID          NOT NULL DEFAULT gen_random_uuid(),
    identifier   VARCHAR       NOT NULL,
    version      VARCHAR       NOT NULL,
    upload_date  CREATION_DATE NOT NULL,
    storage_path VARCHAR       NOT NULL,

    CONSTRAINT pk_artefact PRIMARY KEY (uuid),
    CONSTRAINT un_artefact_identifier_version UNIQUE (identifier, version)
);

CREATE TABLE server
(
    uuid        UUID    NOT NULL DEFAULT gen_random_uuid(),
    environment VARCHAR NOT NULL,
    name        VARCHAR NOT NULL,
    host        VARCHAR NOT NULL,
    memory      INTEGER NOT NULL,
    image       VARCHAR NOT NULL,

    CONSTRAINT pk_servers_uuid PRIMARY KEY (uuid),
    CONSTRAINT un_servers_environment_name UNIQUE (environment, name),
    CONSTRAINT un_server_uuid_host UNIQUE (uuid, host)
);

CREATE TABLE server_cpu_allocation
(
    uuid        UUID     NOT NULL DEFAULT gen_random_uuid(),
    server_uuid UUID     NOT NULL,
    server_host VARCHAR  NOT NULL,
    cpu_core    CPU_CORE NOT NULL,

    CONSTRAINT pk_server_cpu_allocation PRIMARY KEY (uuid),
    CONSTRAINT un_server_cpu_allocation UNIQUE (server_uuid, server_host, cpu_core),
    CONSTRAINT fk_server_cpu_allocation_server FOREIGN KEY (server_uuid, server_host) REFERENCES server (uuid, host)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE server_state
(
    uuid            UUID              NOT NULL DEFAULT gen_random_uuid(),
    server          UUID              NOT NULL,
    artefact        UUID              NOT NULL,
    definition_date CREATION_DATE     NOT NULL,
    type            SERVER_STATE_TYPE NOT NULL,

    CONSTRAINT pk_server_state PRIMARY KEY (uuid),
    CONSTRAINT un_server_state UNIQUE (server, artefact),
    CONSTRAINT fk_server FOREIGN KEY (server) REFERENCES server (uuid),
    CONSTRAINT fk_artefact FOREIGN KEY (artefact) REFERENCES artefact (uuid)
);

CREATE INDEX idx_server_state_non_history_state_server ON server_state (server) WHERE type != 'HISTORY';
CREATE INDEX idx_server_state_non_history_state_artefact ON server_state (artefact) WHERE type != 'HISTORY';
CREATE UNIQUE INDEX idx_server_state_is_uniq ON server_state (server, artefact) WHERE type = 'IS';
CREATE UNIQUE INDEX idx_server_state_target_uniq ON server_state (server, artefact) WHERE type = 'TARGET';
