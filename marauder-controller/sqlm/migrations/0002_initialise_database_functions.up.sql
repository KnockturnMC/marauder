--
-- Function to query the current missmatches in artefacts between the passed servers is and target state.
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
	RETURN QUERY SELECT is_artefact.identifier,
						is_artefact.uuid,
						is_artefact.version,
						target_artefact.uuid,
						target_artefact.version
				 FROM server_state is_state
						  JOIN server_state target_state ON is_state.server = target_state.server
					 AND is_state.type = 'IS'
					 AND target_state.type = 'TARGET'
					 AND is_state.artefact_uuid != target_state.artefact_uuid
					 AND is_state.artefact_identifier = target_state.artefact_identifier
						  JOIN artefact is_artefact on is_artefact.uuid = is_state.artefact_uuid
						  JOIN artefact target_artefact on target_artefact.uuid = target_state.artefact_uuid
				 WHERE is_state.type = 'IS'
				   AND is_state.server = server_uuid;
END
$$ LANGUAGE plpgsql;

--
-- Function to update the current target state of a server in the controller with a new artefact
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
		  AND artefact_identifier = param_artefact_identifier; -- delete current is state, not archived like target one
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
-- Function to query all server states of a server given the state type.
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
