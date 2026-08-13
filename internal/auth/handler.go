package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"store_auth/internal/middleware"
	"store_auth/internal/otp"
)

// Handler exposes HTTP handlers for authentication and user account endpoints.
type Handler struct {
	authService *Service
	isProd      bool
}

// NewHandler constructs an Auth Handler with service dependencies.
func NewHandler(authService *Service, isProd bool) *Handler {
	return &Handler{
		authService: authService,
		isProd:      isProd,
	}
}

// Register processes incoming user registration requests.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if err := validatePassword(req.Password); err != nil {
		details["password"] = err.Error()
	}
	if strings.TrimSpace(req.Name) == "" {
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
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req otp.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if strings.TrimSpace(req.Code) == "" {
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

// Login authenticates credentials and sets an HttpOnly JWT cookie on successful validation.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

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

	usr, token, err := h.authService.Login(r.Context(), req)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteStrictMode,
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    usr.ToUserResponse(),
	})
}

// Logout invalidates the client's session cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteStrictMode,
	})

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// ForgotPassword accepts email address to initiate password reset via OTP code.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

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
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req otp.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	details := make(map[string]string)
	if err := validateEmail(req.Email); err != nil {
		details["email"] = err.Error()
	}
	if strings.TrimSpace(req.Code) == "" {
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

	respondJSON(w, http.StatusOK, user.ToUserResponse())
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
