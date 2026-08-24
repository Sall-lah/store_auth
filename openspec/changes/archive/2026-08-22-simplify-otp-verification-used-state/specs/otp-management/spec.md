## MODIFIED Requirements

### Requirement: OTP secret code lifecycle state
The system SHALL transition the OTP record to `used = true` whenever an OTP is successfully verified, replaced by a resend, or invalidated upon account activation or password reset completion, preventing any replay or reuse of the verification code.

#### Scenario: Marking OTP as used upon successful verification
- **WHEN** an OTP is verified successfully during account activation or password reset
- **THEN** the system updates the record to `used = true` in the database and prevents subsequent verification of the same code

#### Scenario: Invalidation of prior OTPs on re-issuance
- **WHEN** a new OTP is issued due to a resend request, unverified re-registration, or completed password recovery
- **THEN** the system marks all existing pending OTPs for that user and type as `used = true`
