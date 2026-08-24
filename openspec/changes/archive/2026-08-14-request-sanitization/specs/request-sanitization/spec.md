## ADDED Requirements

### Requirement: Global HTTP request body size limiting
The system SHALL reject incoming HTTP request payloads whose size exceeds 1 MB (1,048,576 bytes) by enforcing request body size restrictions before JSON decoding.

#### Scenario: Request body exceeds size limit
- **WHEN** a client sends an HTTP POST request payload larger than 1 MB to any authentication endpoint
- **THEN** the system rejects the request and returns HTTP 400 Bad Request or HTTP 413 Payload Too Large

#### Scenario: Request body within size limit
- **WHEN** a client sends an HTTP POST request payload equal to or smaller than 1 MB
- **THEN** the system processes the request normally

### Requirement: Input string trimming and email normalization
The system SHALL strip leading and trailing whitespace from all incoming string input fields (`email`, `name`, `code`) and normalize all email addresses to lowercase before performing validation, database queries, or user creation.

#### Scenario: Email submitted with mixed case and leading/trailing whitespace
- **WHEN** a user submits an email `" User.Test@Example.COM "` during registration, login, or password reset
- **THEN** the system sanitizes and normalizes the email to `"user.test@example.com"` before performing uniqueness checks, database lookups, or user storage

### Requirement: User display name HTML and control character sanitization
The system SHALL sanitize string input fields such as `name` by stripping HTML elements, script tags, and non-printable control characters to prevent Stored XSS and invalid unicode input.

#### Scenario: Name submitted with HTML script tags
- **WHEN** a user submits a name containing HTML tags such as `<script>alert('xss')</script>John Doe`
- **THEN** the system sanitizes the input to `"John Doe"` before validating and persisting the user record
