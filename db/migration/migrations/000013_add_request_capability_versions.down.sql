BEGIN;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS requested_fhir_version; 
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS requested_fhir_version; 
ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS capability_fhir_version;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS capability_fhir_version;

COMMIT;