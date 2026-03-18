-- Extend the main server table via a path field to the management socket.
-- Enables optional server management queries / calls.
ALTER TABLE server
	ADD COLUMN management_socket_path VARCHAR NOT NULL DEFAULT '';
