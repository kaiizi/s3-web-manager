## ADDED Requirements

### Requirement: List all buckets via Admin API
The system SHALL retrieve the full list of buckets across all owners by calling the Ceph RGW Admin OPS API (`GET /admin/bucket?list`). The response SHALL include bucket name and owner for each bucket. Standard S3 `ListAllMyBuckets` MUST NOT be used as the primary listing mechanism.

#### Scenario: List all buckets
- **WHEN** authenticated user requests `GET /api/buckets`
- **THEN** system returns HTTP 200 with a JSON array of all buckets including `name` and `owner` fields

#### Scenario: Empty bucket list
- **WHEN** no buckets exist in Ceph
- **THEN** system returns HTTP 200 with an empty array

#### Scenario: Ceph Admin API unreachable
- **WHEN** Ceph RGW endpoint is unreachable
- **THEN** system returns HTTP 502 with an error message

### Requirement: Bucket detail with stats
The system SHALL retrieve detailed information for a specific bucket by calling `GET /admin/bucket?bucket=<name>&stats=true`. The response SHALL include size, object count, and owner.

#### Scenario: Get bucket detail
- **WHEN** authenticated user requests `GET /api/buckets/:name`
- **THEN** system returns HTTP 200 with bucket details including `name`, `owner`, `numObjects`, `sizeBytes`

#### Scenario: Bucket not found
- **WHEN** requested bucket name does not exist
- **THEN** system returns HTTP 404 with an error message

### Requirement: Frontend bucket list page
The frontend SHALL display all buckets in a table with columns for name, owner, and actions. Clicking a bucket SHALL navigate to its detail/object browser page.

#### Scenario: Display bucket list
- **WHEN** authenticated user visits `/buckets`
- **THEN** frontend displays a table with all buckets fetched from the backend

#### Scenario: Navigate to bucket detail
- **WHEN** user clicks on a bucket row
- **THEN** frontend navigates to `/buckets/:name`
