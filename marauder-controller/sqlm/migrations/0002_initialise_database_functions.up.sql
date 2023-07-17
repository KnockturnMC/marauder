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
					 AND is_state.artefact != target_state.artefact
						  JOIN artefact is_artefact on is_artefact.uuid = is_state.artefact
						  JOIN artefact target_artefact on target_artefact.uuid = target_state.artefact
				 WHERE is_state.type = 'IS'
				   AND is_state.artefact != target_state.artefact
				   AND is_state.server = server_uuid;
END
$$ LANGUAGE plpgsql;

--
-- Function to update the current target state of a server in the controller with a new artefact
--
CREATE FUNCTION func_update_server_target_state(server_uuid UUID, artefact_uuid UUID)
	RETURNS server_state AS
$$
DECLARE
	inserted_row server_state;
BEGIN
	UPDATE server_state SET type = 'HISTORY' WHERE server = server_uuid AND type = 'TARGET'; -- archive current target state to history


	INSERT
	INTO server_state (server, artefact, definition_date, type)
	VALUES (server_uuid, artefact_uuid, NOW(), 'TARGET')
	RETURNING *
		INTO inserted_row;
	return inserted_row;
END
$$ LANGUAGE plpgsql;
