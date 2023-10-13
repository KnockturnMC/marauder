-- The main table for scheduled lifecycle actions holds the actions, their time of execution and the
-- server they are scheduled against.
CREATE TABLE IF NOT EXISTS scheduled_lifecycle_actions
(
	uuid              UUID      NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	server            UUID      NOT NULL,
	action            VARCHAR   NOT NULL,
	time_of_execution TIMESTAMP NOT NULL,

	CONSTRAINT fk_scheduled_lifecycle_actions_server FOREIGN KEY (server) REFERENCES server (uuid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Simple function that inserts a new scheduled lifecycle action to the table or skips it if an earlier one already exists.
CREATE FUNCTION func_insert_or_join_existing_scheduled_lifecycle_action(
	func_server UUID,
	func_action VARCHAR,
	func_time_of_execution TIMESTAMP
)
	RETURNS scheduled_lifecycle_actions
AS
$$
DECLARE
	return_value scheduled_lifecycle_actions;
BEGIN
	SELECT *
	INTO return_value
	FROM scheduled_lifecycle_actions s
	WHERE s.server = func_server
	  AND s.action = func_action
	  AND s.time_of_execution < func_time_of_execution;

	IF return_value IS NULL THEN
		INSERT INTO scheduled_lifecycle_actions(server, action, time_of_execution)
		VALUES (func_server, func_action, func_time_of_execution)
		RETURNING *
			INTO return_value;
	end if;

	RETURN return_value;
END
$$ LANGUAGE plpgsql;

