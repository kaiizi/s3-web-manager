## ADDED Requirements

### Requirement: List objects with prefix-based directory navigation
The system SHALL list objects in a bucket using `ListObjectsV2` with `prefix` and `delimiter=/` parameters to simulate a directory tree. The response SHALL separately return common prefixes (sub-"directories") and object entries at the current level.

#### Scenario: List root level objects
- **WHEN** authenticated user requests `GET /api/buckets/:name/objects` (no prefix)
- **THEN** system returns HTTP 200 with `{"prefixes": [...], "objects": [...]}`

#### Scenario: List objects under prefix
- **WHEN** authenticated user requests `GET /api/buckets/:name/objects?prefix=photos/`
- **THEN** system returns only prefixes and objects immediately under `photos/`

#### Scenario: Empty prefix
- **WHEN** bucket has no objects
- **THEN** system returns HTTP 200 with `{"prefixes": [], "objects": []}`

### Requirement: Generate Presigned upload URL
The system SHALL generate a Presigned PUT URL for uploading an object directly to Ceph. The URL SHALL be valid for 15 minutes. The backend MUST NOT receive or relay the file data.

#### Scenario: Generate upload URL
- **WHEN** authenticated user requests `POST /api/buckets/:name/objects/upload-url` with body `{"key": "path/to/file.txt", "contentType": "text/plain"}`
- **THEN** system returns HTTP 200 with `{"url": "<presigned-put-url>", "method": "PUT"}`

#### Scenario: Frontend uploads using Presigned URL
- **WHEN** frontend receives the Presigned URL
- **THEN** frontend PUTs the file directly to Ceph using the URL (bypassing Go backend)

#### Scenario: Upload progress display
- **WHEN** user uploads a file using the Presigned URL
- **THEN** frontend displays upload progress percentage using XHR/fetch progress events

### Requirement: Generate Presigned download URL
The system SHALL generate a Presigned GET URL for downloading an object directly from Ceph. The URL SHALL be valid for 15 minutes. The backend MUST NOT stream the file data.

#### Scenario: Generate download URL
- **WHEN** authenticated user requests `GET /api/buckets/:name/objects/download-url?key=path/to/file.txt`
- **THEN** system returns HTTP 200 with `{"url": "<presigned-get-url>"}`

#### Scenario: Frontend triggers download
- **WHEN** frontend receives the Presigned download URL
- **THEN** frontend triggers browser file download via the URL directly

### Requirement: Delete object
The system SHALL delete a specified object from a bucket using the standard S3 `DeleteObject` API.

#### Scenario: Delete existing object
- **WHEN** authenticated user requests `DELETE /api/buckets/:name/objects?key=path/to/file.txt`
- **THEN** system returns HTTP 204 and the object is removed from Ceph

#### Scenario: Delete non-existent object
- **WHEN** key does not exist in the bucket
- **THEN** system returns HTTP 204 (S3 DeleteObject is idempotent)

#### Scenario: Confirm before delete
- **WHEN** user clicks the delete button on an object in the frontend
- **THEN** frontend displays a confirmation dialog before sending the delete request

### Requirement: Frontend object browser
The frontend SHALL display objects and sub-prefixes in a table within the bucket detail page. Users SHALL be able to navigate into sub-prefixes (simulated directories), upload files, download files, and delete files from this view.

#### Scenario: Display objects and prefixes
- **WHEN** user visits `/buckets/:name` or navigates to a sub-prefix
- **THEN** frontend displays a breadcrumb trail of current prefix, a table with sub-prefixes and objects, and action buttons (download, delete) per object

#### Scenario: Navigate into sub-prefix
- **WHEN** user clicks a prefix entry (simulated directory)
- **THEN** frontend updates the current prefix and re-fetches objects for that prefix
