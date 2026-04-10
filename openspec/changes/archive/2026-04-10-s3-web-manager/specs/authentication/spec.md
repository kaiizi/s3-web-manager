## ADDED Requirements

### Requirement: Fixed credential login
The system SHALL authenticate users via a single fixed username and password configured through environment variables (`APP_USERNAME`, `APP_PASSWORD`). On successful authentication, a JWT token (HS256, 24-hour expiry) SHALL be returned. The JWT secret SHALL be read from the `JWT_SECRET` environment variable (minimum 32 characters).

#### Scenario: Successful login
- **WHEN** user submits correct username and password to `POST /api/auth/login`
- **THEN** system returns HTTP 200 with `{"token": "<jwt>"}` in the response body

#### Scenario: Wrong credentials
- **WHEN** user submits incorrect username or password to `POST /api/auth/login`
- **THEN** system returns HTTP 401 with an error message

#### Scenario: Missing credentials
- **WHEN** user submits a request to `POST /api/auth/login` with missing username or password fields
- **THEN** system returns HTTP 400 with a validation error message

### Requirement: JWT middleware protection
All API endpoints except `POST /api/auth/login` SHALL be protected by JWT middleware. Requests without a valid token SHALL be rejected.

#### Scenario: Valid token access
- **WHEN** request includes `Authorization: Bearer <valid-jwt>` header
- **THEN** middleware allows the request to proceed

#### Scenario: Missing token
- **WHEN** request has no `Authorization` header
- **THEN** system returns HTTP 401

#### Scenario: Expired token
- **WHEN** request includes an expired JWT token
- **THEN** system returns HTTP 401

#### Scenario: Invalid token
- **WHEN** request includes a malformed or tampered JWT
- **THEN** system returns HTTP 401

### Requirement: Frontend login page
The frontend SHALL provide a login page as the entry point. Unauthenticated users SHALL be redirected to the login page. After successful login, the token SHALL be stored in localStorage and the user redirected to the bucket list page.

#### Scenario: Login page redirect
- **WHEN** unauthenticated user navigates to any protected route
- **THEN** frontend redirects to `/login`

#### Scenario: Token persistence
- **WHEN** user successfully logs in
- **THEN** JWT token is stored in localStorage and user is redirected to `/buckets`

#### Scenario: Logout
- **WHEN** user clicks logout
- **THEN** token is removed from localStorage and user is redirected to `/login`
