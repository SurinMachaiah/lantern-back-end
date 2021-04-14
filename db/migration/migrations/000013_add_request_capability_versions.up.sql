BEGIN;

DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info;

ALTER TABLE fhir_endpoints_info 
ADD COLUMN requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info 
ADD COLUMN capability_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN capability_fhir_version VARCHAR(500);

CREATE OR REPLACE FUNCTION populate_capability_fhir_version_info() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select capability_statement from fhir_endpoints_info;
        t_row fhir_endpoints_info%ROWTYPE;
        capStatVersion VARCHAR(500);
    BEGIN
        FOR t_row in t_curs LOOP
            SELECT cast(coalesce(nullif(t_row.capability_statement->>'fhirVersion',NULL),'') as varchar(500)) INTO capStatVersion;
            UPDATE fhir_endpoints_info SET requested_fhir_version = '', capability_fhir_version = capStatVersion WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT populate_capability_fhir_version_info();

CREATE OR REPLACE FUNCTION populate_capability_fhir_version_info_history() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select capability_statement from fhir_endpoints_info_history;
        t_row fhir_endpoints_info%ROWTYPE;
        capStatVersion VARCHAR(500);
    BEGIN
        FOR t_row in t_curs LOOP
            SELECT cast(coalesce(nullif(t_row.capability_statement->>'fhirVersion',NULL),'') as varchar(500)) INTO capStatVersion;
            UPDATE fhir_endpoints_info_history SET requested_fhir_version = '', capability_fhir_version = capStatVersion WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT populate_capability_fhir_version_info_history();

-- captures history for the fhir_endpoint_info table
CREATE TRIGGER add_fhir_endpoint_info_history_trigger
AFTER INSERT OR UPDATE OR DELETE on fhir_endpoints_info
FOR EACH ROW
WHEN (current_setting('metadata.setting', 't') IS NULL OR current_setting('metadata.setting', 't') = 'FALSE')
EXECUTE PROCEDURE add_fhir_endpoint_info_history();

COMMIT;