--
-- Function to query the current missmatches in artefacts between the passed servers is and target state.
-- Returns the artefact identifier and the current is artefact and target artefact (both their uuid and their versions).
-- E.g. spellcore on the server might be out of date, running 1.0.0 but should be running 1.0.1.
-- This function would yield "spellcore", the uuid of the 1.0.0 spellcore artefact, 1.0.0, and the uuid of 1.0.1 spellcore and 1.0.1.
--
CREATE FUNCTION func_find_server_target_state_missmatches(server_uuid UUID)
	RETURNS TABLE
			(
				artefact_identifier VARCHAR,
				is_artefact         UUID,
				is_version          VARCHAR,
				target_artefact     UUID,
				target_version      VARCHAR
			)
AS
$$
BEGIN
	RETURN QUERY SELECT COALESCE(target_artefact.identifier, is_artefact.identifier) as identifier,
						is_artefact.uuid,
						is_artefact.version,
						target_artefact.uuid,
						target_artefact.version
				 FROM server_state_target target_state
						  FULL OUTER JOIN server_state_is is_state ON
							 target_state.server = is_state.server
						 AND target_state.artefact_identifier = is_state.artefact_identifier -- Only join with same artefacts
						  LEFT JOIN artefact target_artefact on target_artefact.uuid = target_state.artefact_uuid
						  LEFT JOIN artefact is_artefact on is_artefact.uuid = is_state.artefact_uuid
				 WHERE target_state.artefact_uuid IS DISTINCT FROM is_state.artefact_uuid
				   AND COALESCE(target_state.server, is_state.server) = server_uuid;
END
$$ LANGUAGE plpgsql;

--
-- Function to update the current target state of a server in the controller with a new artefact.
-- This method actively takes care of moving the TARGET state to the HISTORY and deleting the existing state if the target state is TARGET or IS.
--
CREATE FUNCTION func_create_server_state(
	param_server_uuid UUID,
	param_artefact_identifier VARCHAR,
	param_artefact_uuid UUID,
	param_state_type SERVER_STATE_TYPE
)
	RETURNS server_state AS
$$
DECLARE
	inserted_row server_state;
BEGIN
	IF param_state_type = 'TARGET' then
		UPDATE server_state
		SET type = 'HISTORY'
		WHERE server = param_server_uuid
		  AND artefact_identifier = param_artefact_identifier
		  AND type = param_state_type; -- archive current target state to history
	end if;
	IF param_state_type = 'IS' then
		DELETE
		FROM server_state
		WHERE server = param_server_uuid
		  AND type = param_state_type
		  AND artefact_identifier = param_artefact_identifier; -- delete current is state, not archived like target one.
	end if;

	INSERT
	INTO server_state (server, artefact_identifier, artefact_uuid, definition_date, type)
	VALUES (param_server_uuid, param_artefact_identifier, param_artefact_uuid, NOW(), param_state_type)
	RETURNING *
		INTO inserted_row;
	return inserted_row;
END
$$ LANGUAGE plpgsql;

--
-- Function to query all artefacts under a servers given server state.
--
CREATE FUNCTION func_find_server_artefacts_by_state(server_uuid UUID, state SERVER_STATE_TYPE)
	RETURNS SETOF artefact
AS
$$
BEGIN
	RETURN QUERY SELECT artefact.*
				 FROM server_state
						  JOIN artefact ON server_state.artefact_uuid = artefact.uuid
					 AND server_state.server = server_uuid
					 AND server_state.type = state;
END
$$ LANGUAGE plpgsql;

--
-- Function to query all artefacts older than the passed timestamp that are not either a TARGET or IS state.
--
CREATE FUNCTION func_find_historic_artefacts_older_than(date TIMESTAMP)
	RETURNS SETOF artefact
AS
$$
BEGIN
	RETURN QUERY SELECT artefact.*
				 FROM artefact
						  LEFT JOIN server_state s ON artefact.uuid = s.artefact_uuid
				 WHERE artefact.upload_date < date
				   AND (s.type IS NULL OR s.type = 'HISTORY');
END
$$ LANGUAGE plpgsql;
