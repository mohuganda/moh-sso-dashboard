package auth

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

const (
	cookiePKCEVerifier = "pkce_verifier"
	cookieOAuthState   = "oauth_state"
	cookieAccessToken  = "access_token"
	cookieRefreshToken = "refresh_token"
	cookieIDToken      = "id_token"
)

type Handler struct {
	authService         service.AuthService
	auditService        *service.AuditService
	notificationService service.NotificationsService
	config              *config.Config
}

func NewHandler(
	authService service.AuthService,
	auditService *service.AuditService,
	notificationService service.NotificationsService,
	config *config.Config,
) *Handler {
	return &Handler{
		authService:         authService,
		auditService:        auditService,
		notificationService: notificationService,
		config:              config,
	}
}

// ----------------------------------------------------
// LOGIN (redirect → Keycloak with PKCE + state)
// ----------------------------------------------------
func (h *Handler) HandleAuthLogin(c *gin.Context) {
	_ = h.auditService.LoginInitiated(
		c.Request.Context(),
		c.ClientIP(),
		c.Request.UserAgent(),
		h.config.KeycloakWebClientID,
	)

	codeVerifier, _ := c.Cookie(cookiePKCEVerifier)
	state, _ := c.Cookie(cookieOAuthState)

	if codeVerifier == "" || state == "" {
		codeVerifier = utils.GenerateCodeVerifier()
		state = uuid.NewString()

		h.setCookie(c, cookiePKCEVerifier, codeVerifier, 300, true)
		h.setCookie(c, cookieOAuthState, state, 300, true)
	}

	codeChallenge := utils.GenerateCodeChallenge(codeVerifier)

	keycloakBaseURL := strings.TrimRight(h.config.KeycloakExternalURL, "/")

	authURL, err := url.Parse(
		keycloakBaseURL +
			"/realms/" + h.config.KeycloakRealm +
			"/protocol/openid-connect/auth",
	)
	if err != nil {
		log.Println("[AUTH LOGIN] failed to parse Keycloak URL:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	q := authURL.Query()
	q.Set("client_id", h.config.KeycloakWebClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", h.config.KeycloakRedirectURI)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")

	authURL.RawQuery = q.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

// ----------------------------------------------------
// GET CURRENT USER (API)
// ----------------------------------------------------
func (h *Handler) HandleAuthGetMe(c *gin.Context) {
	accessToken, err := c.Cookie(cookieAccessToken)
	if err != nil || accessToken == "" {
		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrUnauthorized.Code,
			apierror.ErrUnauthorized.Message,
		)
		return
	}

	user, err := h.authService.GetMe(accessToken)
	if err != nil {
		log.Printf(
			"[AUTH ME] failed: ip=%s user_agent=%s err=%v",
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrTokenInvalid.Code,
			"Invalid or expired token",
		)
		return
	}

	userID := utils.ExtractUserIDFromJWT(accessToken)

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		"auth.get_me",
		map[string]interface{}{
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	response.OK(c, http.StatusOK, gin.H{
		"user": user,
	})
}

// ----------------------------------------------------
// OIDC CALLBACK (PKCE + state verification)
// ----------------------------------------------------
func (h *Handler) HandleAuthCallback(c *gin.Context) {
	// Keycloak may redirect back with an error instead of a code.
	if kcErr := c.Query("error"); kcErr != "" {
		h.auditLoginFailure(c)

		log.Printf(
			"[AUTH CALLBACK] keycloak error: error=%s description=%s ip=%s user_agent=%s",
			kcErr,
			c.Query("error_description"),
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":             "keycloak authentication error",
			"keycloak_error":    kcErr,
			"error_description": c.Query("error_description"),
		})
		return
	}

	code := c.Query("code")
	if code == "" {
		h.auditLoginFailure(c)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing authorization code",
		})
		return
	}

	// 🔐 Validate OAuth state.
	returnedState := c.Query("state")
	cookieState, err := c.Cookie(cookieOAuthState)
	if err != nil || cookieState == "" || returnedState == "" || returnedState != cookieState {
		details := "oauth state mismatch"
		if err != nil {
			details = err.Error()
		}

		log.Printf(
			"[AUTH CALLBACK] invalid oauth state: returned_state_present=%v cookie_state_present=%v ip=%s user_agent=%s err=%v",
			returnedState != "",
			cookieState != "",
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		// If the user refreshed an already-successful callback URL,
		// oauth_state may be gone because we cleared it after login.
		// If auth cookies exist, redirect to the frontend instead of showing JSON.
		if h.hasAnyAuthCookie(c) {
			c.Redirect(http.StatusTemporaryRedirect, h.defaultFrontendRedirect())
			return
		}

		h.auditLoginFailure(c)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid oauth state",
			"details": details,
		})
		return
	}

	// 🔐 Read PKCE verifier.
	codeVerifier, err := c.Cookie(cookiePKCEVerifier)
	if err != nil || codeVerifier == "" {
		h.auditLoginFailure(c)

		details := "pkce verifier cookie not found"
		if err != nil {
			details = err.Error()
		}

		log.Printf(
			"[AUTH CALLBACK] missing pkce verifier: ip=%s user_agent=%s err=%v",
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "missing pkce verifier",
			"details": details,
		})
		return
	}

	tokens, err := h.authService.ProcessAuthCode(code, codeVerifier)
	if err != nil {
		log.Printf(
			"[AUTH CALLBACK] token exchange failed: ip=%s user_agent=%s err=%v",
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		// Authorization codes are one-time use.
		// If the user refreshes the callback URL, Keycloak returns invalid_grant / Code not valid.
		// In that case, if the user already has a session cookie, redirect them to the frontend.
		if h.isCodeAlreadyUsedError(err) {
			if _, cookieErr := c.Cookie(cookieAccessToken); cookieErr == nil {
				c.Redirect(http.StatusTemporaryRedirect, h.defaultFrontendRedirect())
				return
			}

			if _, cookieErr := c.Cookie(cookieRefreshToken); cookieErr == nil {
				c.Redirect(http.StatusTemporaryRedirect, h.defaultFrontendRedirect())
				return
			}
		}

		h.notificationService.NotifyLoginFailed(
			c.Request.Context(),
			h.config.KeycloakWebClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		h.auditLoginFailure(c)

		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{
				"error":   "authentication failed",
				"details": err.Error(),
			},
		)
		return
	}

	if tokens == nil || tokens.AccessToken == "" {
		h.auditLoginFailure(c)

		log.Printf(
			"[AUTH CALLBACK] empty token response: ip=%s user_agent=%s",
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{
				"error":   "authentication failed",
				"details": "empty access token",
			},
		)
		return
	}

	// Clear one-time cookies after successful exchange.
	h.clearCookie(c, cookiePKCEVerifier, true)
	h.clearCookie(c, cookieOAuthState, true)

	userID := utils.ExtractUserIDFromJWT(tokens.AccessToken)

	_ = h.auditService.LoginResult(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		true,
		h.config.KeycloakWebClientID,
		c.ClientIP(),
		c.Request.UserAgent(),
		"",
		"",
	)

	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)

	if tokens.RefreshToken != "" && tokens.RefreshExpiresIn > 0 {
		h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)
	}

	if tokens.IDToken != "" {
		h.setSecureIDTokenCookie(c, tokens.IDToken, int(tokens.ExpiresIn))
	}

	// ------------------------------------------------------------------
	// Role-based redirect (authoritative)
	// ------------------------------------------------------------------
	isAdmin := utils.TokenHasRealmRole(tokens.AccessToken, "admin")
	isUser := utils.TokenHasRealmRole(tokens.AccessToken, "user")

	baseURL := strings.TrimRight(h.config.FrontendBaseURL, "/")

	redirectURL := baseURL + "/"
	if isAdmin {
		redirectURL = baseURL + "/admin/home"
	} else if isUser {
		redirectURL = baseURL + "/apps/news"
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// ----------------------------------------------------
// REFRESH TOKEN (API)
// ----------------------------------------------------
func (h *Handler) HandleAuthRefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(cookieRefreshToken)
	if err != nil || refreshToken == "" {
		_ = h.auditService.TokenRefresh(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrTokenExpired.Code,
			"Session expired",
		)
		return
	}

	tokens, err := h.authService.GetAccessToken(refreshToken)
	if err != nil {
		log.Printf(
			"[AUTH REFRESH] failed: ip=%s user_agent=%s err=%v",
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		h.notificationService.NotifyTokenRefreshFailed(
			c.Request.Context(),
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		_ = h.auditService.TokenRefresh(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		h.clearAuthCookies(c)

		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrTokenInvalid.Code,
			"Invalid or expired refresh token",
		)
		return
	}

	if tokens == nil || tokens.AccessToken == "" {
		log.Printf("[AUTH REFRESH] invalid token response from Keycloak")

		h.clearAuthCookies(c)

		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrTokenInvalid.Code,
			"Invalid token response",
		)
		return
	}

	userID := utils.ExtractUserIDFromJWT(tokens.AccessToken)

	_ = h.auditService.TokenRefresh(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		true,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)

	// Important:
	// With refresh token rotation, Keycloak may return a new refresh token.
	// Only overwrite the cookie if a new refresh token was actually returned.
	if tokens.RefreshToken != "" && tokens.RefreshExpiresIn > 0 {
		h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)
	}

	if tokens.IDToken != "" {
		h.setSecureIDTokenCookie(c, tokens.IDToken, int(tokens.ExpiresIn))
	}

	response.OK(c, http.StatusOK, gin.H{
		"expires_in": tokens.ExpiresIn,
		"cookie_debug": gin.H{
			"environment":        h.environment(),
			"secure":             h.cookieSecure(),
			"same_site":          h.cookieSameSiteString(),
			"domain":             h.cookieDomain(),
			"has_access_token":   tokens.AccessToken != "",
			"has_refresh_token":  tokens.RefreshToken != "",
			"refresh_expires_in": tokens.RefreshExpiresIn,
			"frontend_base_url":  h.config.FrontendBaseURL,
			"keycloak_redirect":  h.config.KeycloakRedirectURI,
			"keycloak_external":  h.config.KeycloakExternalURL,
			"keycloak_internal":  h.config.KeycloakTokenBaseURL(),
		},
	})
}

// ----------------------------------------------------
// LOGOUT
// ----------------------------------------------------
func (h *Handler) HandleAuthLogout(c *gin.Context) {
	ctx := c.Request.Context()

	userID := utils.ToNullUUID(c.GetString("user_id"))

	_ = h.auditService.Logout(
		ctx,
		userID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	refreshToken, _ := c.Cookie(cookieRefreshToken)
	idTokenHint, _ := c.Cookie(cookieIDToken)

	if refreshToken != "" {
		if err := h.authService.LogOut(refreshToken); err != nil {
			log.Printf("[AUTH LOGOUT] token revocation failed: %v", err)
		}
	}

	redirectURL := h.authService.GetLogoutURL(
		idTokenHint,
		h.config.FrontendBaseURL,
	)

	h.clearAuthCookies(c)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// ----------------------------------------------------
// AUDIT HELPERS
// ----------------------------------------------------
func (h *Handler) auditLoginFailure(c *gin.Context) {
	_ = h.auditService.LoginResult(
		c.Request.Context(),
		uuid.NullUUID{},
		false,
		h.config.KeycloakWebClientID,
		c.ClientIP(),
		c.Request.UserAgent(),
		"",
		"",
	)
}

// ----------------------------------------------------
// CONFIG / ENV HELPERS
// ----------------------------------------------------
func (h *Handler) environment() string {
	if h == nil || h.config == nil {
		return "development"
	}

	env := strings.TrimSpace(h.config.Environment)
	if env == "" {
		return "development"
	}

	return strings.ToLower(env)
}

func (h *Handler) cookieSecure() bool {
	if h == nil || h.config == nil {
		return false
	}

	return h.config.CookieSecure()
}

func (h *Handler) cookieDomain() string {
	if h == nil || h.config == nil {
		return ""
	}

	return h.config.CookieDomainValue()
}

func (h *Handler) cookieSameSite() http.SameSite {
	// Config currently returns "Lax", but we keep this conversion here
	// so future config changes do not require touching all cookie setters.
	if h == nil || h.config == nil {
		return http.SameSiteLaxMode
	}

	switch strings.ToLower(strings.TrimSpace(h.config.CookieSameSiteMode())) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *Handler) cookieSameSiteString() string {
	switch h.cookieSameSite() {
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return "Default"
	}
}

// ----------------------------------------------------
// COOKIE HELPERS
// ----------------------------------------------------
func (h *Handler) setCookie(
	c *gin.Context,
	name string,
	value string,
	maxAge int,
	httpOnly bool,
) {
	if value == "" || maxAge <= 0 {
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   h.cookieDomain(),
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: httpOnly,
		Secure:   h.cookieSecure(),
		SameSite: h.cookieSameSite(),
	})
}

func (h *Handler) setSecureAccessTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
	if token == "" || maxAge <= 0 {
		return
	}

	h.setCookie(c, cookieAccessToken, token, int(maxAge), true)
}

func (h *Handler) setSecureRefreshTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
	if token == "" || maxAge <= 0 {
		return
	}

	h.setCookie(c, cookieRefreshToken, token, int(maxAge), true)
}

func (h *Handler) setSecureIDTokenCookie(
	c *gin.Context,
	token string,
	expiresIn int,
) {
	if token == "" || expiresIn <= 0 {
		return
	}

	h.setCookie(c, cookieIDToken, token, expiresIn, true)
}

func (h *Handler) clearCookie(
	c *gin.Context,
	name string,
	httpOnly bool,
) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   h.cookieDomain(),
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: httpOnly,
		Secure:   h.cookieSecure(),
		SameSite: h.cookieSameSite(),
	})
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	h.clearCookie(c, cookieAccessToken, true)
	h.clearCookie(c, cookieRefreshToken, true)
	h.clearCookie(c, cookieIDToken, true)
	h.clearCookie(c, cookiePKCEVerifier, true)
	h.clearCookie(c, cookieOAuthState, true)
}

func (h *Handler) isCodeAlreadyUsedError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "invalid_grant") ||
		strings.Contains(msg, "code not valid") ||
		strings.Contains(msg, "code already used")
}

func (h *Handler) hasAnyAuthCookie(c *gin.Context) bool {
	if c == nil {
		return false
	}

	if token, err := c.Cookie(cookieAccessToken); err == nil && token != "" {
		return true
	}

	if token, err := c.Cookie(cookieRefreshToken); err == nil && token != "" {
		return true
	}

	return false
}

func (h *Handler) defaultFrontendRedirect() string {
	if h == nil || h.config == nil {
		return "/"
	}

	baseURL := strings.TrimRight(h.config.FrontendBaseURL, "/")
	if baseURL == "" {
		return "/"
	}

	return baseURL + "/"
}
