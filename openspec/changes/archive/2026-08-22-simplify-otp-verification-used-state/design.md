# Design: OTP Verification State Lifecycle Simplification

## Context

In `store_auth`, One-Time Passwords (OTPs) govern user account activation and self-service password recovery workflows.

When a user submits a valid 6-digit numeric OTP code (`POST /api/auth/verify-otp`), the service verifies:
1. The OTP belongs to the user and matching flow type (`REGISTRATION` vs `PASSWORD_RESET`).
2. The OTP is not expired (`time.Now().Before(expiresAt)`).
3. The OTP has not exceeded 5 invalid submission attempts (`attempts < 5`).
4. The OTP is not already consumed (`used == false`).

Once verified, the service marks the OTP record as consumed and activates the user account.

## Root Cause Analysis

In `internal/otp/repository.go`:
```go
func (r *Repository) MarkOTPUsedAndScrubCode(ctx context.Context, id string) error {
	_, err := r.client.OTPCode.FindUnique(
		db.OTPCode.ID.Equals(id),
	).Update(
		db.OTPCode.Used.Set(true),
		db.OTPCode.Code.SetOptional(nil), // <--- Prisma Query Engine failure
	).Exec(ctx)
```
Prisma Client Go encounters a query engine serialization error when attempting to set an optional string to `nil` in an update query on PostgreSQL.

## Architectural Decision

Instead of attempting to mutate the `code` column to `nil`, we adopt **Approach 1**: Transition the record's boolean state to `used = true`.

```
                      OTP STATE LIFECYCLE
                      ═══════════════════

                       [ Generate OTP ]
                              │
                              ▼
                     ┌──────────────────┐
                     │   used = false   │
                     │   attempts = 0   │
                     │   TTL = 5 mins   │
                     └────────┬─────────┘
                              │
             ┌────────────────┴────────────────┐
     [ Code != input ]                 [ Code == input ]
             │                                 │
             ▼                                 ▼
   ┌────────────────────┐            ┌────────────────────┐
   │    attempts += 1   │            │    used = true     │
   │ (reject if >= 5)   │            │ (account activated)│
   └────────────────────┘            └─────────┬──────────┘
                                               │
                                               ▼
                                     ┌────────────────────┐
                                     │  Replay Blocked:   │
                                     │  • Ignored by query│
                                     │    (used == false) │
                                     │  • 400 Already Used│
                                     └────────────────────┘
```

### Security & Invariant Analysis

1. **Replay Prevention**: `FindLatestOTPByUserAndType` queries `db.OTPCode.Used.Equals(false)`. Once `used = true`, subsequent queries will not return the code, and `VerifyOTP` explicitly returns `ErrOTPAlreadyUsed` (400) if accessed.
2. **Brute-Force Protection**: The 5-attempt lockout counter (`attempts >= 5`) and 5-minute TTL (`time.Now().After(expiresAt)`) remain fully enforced.
3. **Audit History**: Keeping the code string intact with `used = true` retains complete audit traceability in the PostgreSQL table, recording exactly when codes were created, attempted, and verified.

## Implementation Details

1. **`internal/otp/repository.go`**:
   - Refactor `MarkOTPUsedAndScrubCode` into `MarkOTPUsed(ctx, id)`:
     ```go
     func (r *Repository) MarkOTPUsed(ctx context.Context, id string) error {
         _, err := r.client.OTPCode.FindUnique(
             db.OTPCode.ID.Equals(id),
         ).Update(
             db.OTPCode.Used.Set(true),
         ).Exec(ctx)
         if err != nil {
             if errors.Is(err, db.ErrNotFound) {
                 return ErrOTPNotFound
             }
             return fmt.Errorf("failed to mark otp as used: %w", err)
         }
         return nil
     }
     ```
   - Refactor `InvalidateOTPsByUserAndType(ctx, userID, otpType)` to only update `db.OTPCode.Used.Set(true)`.

2. **`internal/otp/otp_test.go`**:
   - Ensure unit tests verify that `MarkOTPUsed` sets `used == true` without expecting `code` to become nil or error out.
