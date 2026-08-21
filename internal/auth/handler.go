package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"store_auth/internal/middleware"
	"store_auth/internal/otp"
	"store_auth/internal/sanitizer"
)

// Handler exposes HTTP handlers for authentication and user account endpoints.
type Handler struct {
	authService *Service
	isProd      bool
}

// NewHandler constructs an Auth Handler with service dependencies.
// Why: Provides dependency injection of auth service logic and environment configuration.
func NewHandler(authService *Service, isProd bool) *Handler {
	return &Handler{
		authService: authService,
		isProd:      isProd,
	}
}

// Register processes incoming user registration requests.
// Why: Creates a new inactive user account and dispatches an OTP verification code.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)
	req.Name = sanitizer.SanitizeName(req.Name)

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if err := validatePassword(req.Password); err != nil {
		details["password"] = err.Error()
	}
	if req.Name == "" {
		details["name"] = "Name is required"
	}

	if len(details) > 0 {
		respondWithError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	if err := h.authService.Register(r.Context(), req); err != nil {
		if errors.Is(err, ErrEmailTaken) {
			respondWithError(w, http.StatusConflict, err.Error(), nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to register user", nil)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"message": "Registration successful. Please verify your OTP to activate your account.",
	})
}

// VerifyOTP handles OTP verification for account activation.
// Why: Confirms ownership of user email address and activates account for subsequent logins.
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req otp.VerifyOTPRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)
	req.Code = sanitizer.SanitizeCode(req.Code)

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if req.Code == "" {
		details["code"] = "OTP code is required"
	}
	if len(details) > 0 {
		respondWithError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	if err := h.authService.VerifyRegistrationOTP(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, otp.ErrOTPInvalidCode):
			respondWithError(w, http.StatusBadRequest, "Invalid OTP code", nil)
		case errors.Is(err, otp.ErrOTPExpired):
			respondWithError(w, http.StatusGone, "OTP code has expired", nil)
		case errors.Is(err, otp.ErrOTPMaxAttempts):
			respondWithError(w, http.StatusTooManyRequests, "Maximum OTP verification attempts exceeded", nil)
		case errors.Is(err, otp.ErrOTPAlreadyUsed):
			respondWithError(w, http.StatusBadRequest, "OTP code already used", nil)
		default:
			respondWithError(w, http.StatusInternalServerError, "OTP verification failed", nil)
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Account verified successfully. You can now log in.",
	})
}

// ResendOTP processes requests to dispatch a fresh verification OTP code.
// Why: Enables users to request a new code when original codes expire or are lost, while guarding against enumeration.
func (h *Handler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req ResendOTPRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)
	req.Type = strings.TrimSpace(req.Type)

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if req.Type == "" {
		details["type"] = "OTP type is required (registration or password_reset)"
	} else if strings.ToLower(req.Type) != "registration" && strings.ToLower(req.Type) != "password_reset" {
		details["type"] = "Invalid OTP type, must be 'registration' or 'password_reset'"
	}

	if len(details) > 0 {
		respondWithError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	if err := h.authService.ResendOTP(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrAccountAlreadyActive):
			respondWithError(w, http.StatusBadRequest, "Account is already verified. Please log in.", nil)
		case errors.Is(err, ErrUserNotFound):
			respondWithError(w, http.StatusBadRequest, "User account not found", nil)
		case errors.Is(err, ErrInvalidOTPType):
			respondWithError(w, http.StatusBadRequest, "Invalid OTP type", nil)
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to resend OTP", nil)
		}
		return
	}

	msg := "A new verification code has been sent to your email."
	if strings.ToLower(req.Type) == "password_reset" {
		msg = "If the email is registered, a password reset code has been sent."
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": msg,
	})
}

// Login authenticates credentials and sets HttpOnly access and refresh token cookies on successful validation.
// Why: Grants authenticated session access via secure HTTP cookies after credential verification.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if req.Password == "" {
		details["password"] = "Password is required"
	}
	if len(details) > 0 {
		respondWithError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	usr, accessToken, refreshToken, err := h.authService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		case errors.Is(err, ErrAccountInactive):
			respondWithError(w, http.StatusForbidden, "Account is not active. Please complete OTP verification.", nil)
		default:
			respondWithError(w, http.StatusInternalServerError, "Login failed", nil)
		}
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    usr.ToUserResponse(),
	})
}

// Refresh handles access token renewal and refresh token rotation.
// Why: Keeps user session active without prompting for credentials while rotating refresh tokens.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		h.clearAuthCookies(w)
		respondWithError(w, http.StatusUnauthorized, "Refresh token cookie missing or empty", nil)
		return
	}

	if h.authService == nil {
		h.clearAuthCookies(w)
		respondWithError(w, http.StatusInternalServerError, "Auth service unavailable", nil)
		return
	}

	usr, newAccessToken, newRefreshToken, err := h.authService.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		h.clearAuthCookies(w)
		switch {
		case errors.Is(err, ErrAccountInactive):
			respondWithError(w, http.StatusForbidden, "Account is not active", nil)
		case errors.Is(err, ErrRefreshTokenExpired):
			respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", nil)
		case errors.Is(err, ErrRefreshTokenReused):
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", nil)
		case errors.Is(err, ErrInvalidRefreshToken):
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", nil)
		default:
			respondWithError(w, http.StatusUnauthorized, "Failed to refresh token", nil)
		}
		return
	}

	h.setAuthCookies(w, newAccessToken, newRefreshToken)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Token refreshed successfully",
		"user":    usr.ToUserResponse(),
	})
}

// Logout invalidates the client's session cookies and revokes the active refresh token.
// Why: Terminates user session on both client and database by clearing cookies and revoking stored refresh token.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.authService != nil {
		if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
			_ = h.authService.RevokeRefreshToken(r.Context(), cookie.Value)
		}
	}

	h.clearAuthCookies(w)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth/refresh",
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

// ForgotPassword accepts email address to initiate password reset via OTP code.
// Why: Dispatches a password reset OTP code if the account is registered without disclosing user existence.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)

	if err := validateEmail(req.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, "Validation failed", map[string]string{"email": err.Error()})
		return
	}

	_ = h.authService.ForgotPassword(r.Context(), req.Email)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists for this email, a password reset code has been sent.",
	})
}

// ResetPassword validates OTP and updates the user's password.
// Why: Allows account recovery by resetting password using a verified OTP code.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req otp.ResetPasswordRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Email = sanitizer.NormalizeEmail(req.Email)
	req.Code = sanitizer.SanitizeCode(req.Code)

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if req.Code == "" {
		details["code"] = "OTP code is required"
	}
	if err := validatePassword(req.NewPassword); err != nil {
		details["new_password"] = err.Error()
	}
	if len(details) > 0 {
		respondWithError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, otp.ErrOTPInvalidCode):
			respondWithError(w, http.StatusBadRequest, "Invalid OTP code", nil)
		case errors.Is(err, otp.ErrOTPExpired):
			respondWithError(w, http.StatusGone, "OTP code has expired", nil)
		case errors.Is(err, otp.ErrOTPMaxAttempts):
			respondWithError(w, http.StatusTooManyRequests, "Maximum OTP verification attempts exceeded", nil)
		default:
			respondWithError(w, http.StatusInternalServerError, "Password reset failed", nil)
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully. You can now log in with your new password.",
	})
}

// GetMe returns current authenticated user profile using token context claims.
// Why: Provides client applications with identity and profile details of the currently authenticated user.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaimsFromContext(r.Context())
	if !ok || claims == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), claims.Subject)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User profile not found", nil)
		return
	}

	w.Header().Set("X-User-Id", user.ID)
	w.Header().Set("X-User-Role", string(user.Role))
	w.Header().Set("X-User-Email", user.Email)

	respondJSON(w, http.StatusOK, user.ToUserResponse())
}

// decodeJSONBody limits request payload size and unmarshals JSON body.
// Why: Prevents memory exhaustion attacks by enforcing max body size while normalizing JSON parsing error responses.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	middleware.LimitRequestBody(w, r, middleware.DefaultMaxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "Request payload exceeds maximum allowed size", nil)
			return false
		}
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return false
	}
	return true
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("Email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("Invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Details: details,
	})
}
