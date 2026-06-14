DROP FUNCTION func_find_historic_artefacts_older_than;
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
	             WHERE artefact.upload_date < date
		           AND artefact.uuid NOT IN (SELECT server_state.artefact_uuid
		                                     FROM server_state
		                                     WHERE type != 'HISTORY');
END
$$ LANGUAGE plpgsql;
