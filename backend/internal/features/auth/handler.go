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

	// Reuse existing login-flow cookies if present.
	// This prevents double login redirects from overwriting oauth_state/pkce_verifier.
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
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = c.GetHeader("Cf-Ray")
	}
	if requestID == "" {
		requestID = "-"
	}

	log.Printf(
		"[AUTH CALLBACK] started: request_id=%s method=%s path=%s raw_query_present=%v ip=%s user_agent=%s",
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery != "",
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	log.Printf(
		"[AUTH CALLBACK] query summary: request_id=%s state_present=%v code_present=%v session_state_present=%v issuer_present=%v error_present=%v",
		requestID,
		c.Query("state") != "",
		c.Query("code") != "",
		c.Query("session_state") != "",
		c.Query("iss") != "",
		c.Query("error") != "",
	)

	// Keycloak may redirect back with an error instead of a code.
	if kcErr := c.Query("error"); kcErr != "" {
		h.auditLoginFailure(c)

		redirectURL := h.defaultFrontendRedirect()

		log.Printf(
			"[AUTH CALLBACK] keycloak error: request_id=%s error=%s description=%s redirect=%s ip=%s user_agent=%s",
			requestID,
			kcErr,
			c.Query("error_description"),
			redirectURL,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	code := c.Query("code")
	if code == "" {
		h.auditLoginFailure(c)

		redirectURL := h.defaultFrontendRedirect()

		log.Printf(
			"[AUTH CALLBACK] missing authorization code: request_id=%s redirect=%s ip=%s user_agent=%s",
			requestID,
			redirectURL,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	log.Printf(
		"[AUTH CALLBACK] authorization code received: request_id=%s code_present=%v",
		requestID,
		code != "",
	)

	// 🔐 Validate OAuth state.
	returnedState := c.Query("state")

	log.Printf(
		"[AUTH CALLBACK] reading oauth state cookie: request_id=%s returned_state_present=%v",
		requestID,
		returnedState != "",
	)

	cookieState, err := c.Cookie(cookieOAuthState)
	if err != nil || cookieState == "" || returnedState == "" || returnedState != cookieState {
		details := "oauth state mismatch"
		if err != nil {
			details = err.Error()
		}

		redirectURL := h.defaultFrontendRedirect()

		log.Printf(
			"[AUTH CALLBACK] invalid oauth state: request_id=%s returned_state_present=%v cookie_state_present=%v states_match=%v redirect=%s ip=%s user_agent=%s err=%v details=%s",
			requestID,
			returnedState != "",
			cookieState != "",
			returnedState != "" && cookieState != "" && returnedState == cookieState,
			redirectURL,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
			details,
		)

		// Stale callback, expired login attempt, refreshed callback URL,
		// or OAuth cookies cleared by the browser.
		// Do not notify admins for this normal browser/OAuth behavior.
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	log.Printf(
		"[AUTH CALLBACK] oauth state validated: request_id=%s",
		requestID,
	)

	// 🔐 Read PKCE verifier.
	log.Printf(
		"[AUTH CALLBACK] reading pkce verifier cookie: request_id=%s",
		requestID,
	)

	codeVerifier, err := c.Cookie(cookiePKCEVerifier)
	if err != nil || codeVerifier == "" {
		details := "pkce verifier cookie not found"
		if err != nil {
			details = err.Error()
		}

		redirectURL := h.defaultFrontendRedirect()

		log.Printf(
			"[AUTH CALLBACK] missing pkce verifier: request_id=%s verifier_present=%v redirect=%s ip=%s user_agent=%s err=%v details=%s",
			requestID,
			codeVerifier != "",
			redirectURL,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
			details,
		)

		// Same as state failure: usually stale/expired callback.
		// Do not notify admins.
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	log.Printf(
		"[AUTH CALLBACK] pkce verifier found: request_id=%s verifier_present=%v",
		requestID,
		codeVerifier != "",
	)

	log.Printf(
		"[AUTH CALLBACK] starting token exchange: request_id=%s client_id=%s ip=%s user_agent=%s",
		requestID,
		h.config.KeycloakWebClientID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	tokens, err := h.authService.ProcessAuthCode(code, codeVerifier)
	if err != nil {
		log.Printf(
			"[AUTH CALLBACK] token exchange failed: request_id=%s ip=%s user_agent=%s err=%v",
			requestID,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		// Authorization codes are one-time use.
		// If the user refreshes the callback URL after a successful exchange,
		// Keycloak returns invalid_grant / Code not valid.
		if h.isCodeAlreadyUsedError(err) {
			redirectURL := h.defaultFrontendRedirect()
			hasAuthCookie := h.hasAnyAuthCookie(c)

			log.Printf(
				"[AUTH CALLBACK] authorization code already used: request_id=%s has_auth_cookie=%v redirect=%s ip=%s user_agent=%s",
				requestID,
				hasAuthCookie,
				redirectURL,
				c.ClientIP(),
				c.Request.UserAgent(),
			)

			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}

		log.Printf(
			"[AUTH CALLBACK] notifying login failure: request_id=%s client_id=%s",
			requestID,
			h.config.KeycloakWebClientID,
		)

		h.notificationService.NotifyLoginFailed(
			c.Request.Context(),
			h.config.KeycloakWebClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		log.Printf(
			"[AUTH CALLBACK] writing login failure audit: request_id=%s",
			requestID,
		)

		h.auditLoginFailure(c)

		log.Printf(
			"[AUTH CALLBACK] returning unauthorized: request_id=%s reason=token_exchange_failed",
			requestID,
		)

		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{
				"error":   "authentication failed",
				"details": err.Error(),
			},
		)
		return
	}

	log.Printf(
		"[AUTH CALLBACK] token exchange completed: request_id=%s tokens_nil=%v access_token_present=%v refresh_token_present=%v id_token_present=%v expires_in=%d refresh_expires_in=%d",
		requestID,
		tokens == nil,
		tokens != nil && tokens.AccessToken != "",
		tokens != nil && tokens.RefreshToken != "",
		tokens != nil && tokens.IDToken != "",
		func() int {
			if tokens == nil {
				return 0
			}
			return int(tokens.ExpiresIn)
		}(),
		func() int {
			if tokens == nil {
				return 0
			}
			return int(tokens.RefreshExpiresIn)
		}(),
	)

	if tokens == nil || tokens.AccessToken == "" {
		h.auditLoginFailure(c)

		log.Printf(
			"[AUTH CALLBACK] empty token response: request_id=%s tokens_nil=%v ip=%s user_agent=%s",
			requestID,
			tokens == nil,
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

	// Clear only OAuth one-time cookies after successful exchange.
	// Do not clear session cookies here.
	log.Printf(
		"[AUTH CALLBACK] clearing oauth one-time cookies: request_id=%s",
		requestID,
	)

	h.clearOAuthCookies(c)

	userID := utils.ExtractUserIDFromJWT(tokens.AccessToken)

	log.Printf(
		"[AUTH CALLBACK] extracted user from access token: request_id=%s user_id_present=%v",
		requestID,
		userID != "",
	)

	log.Printf(
		"[AUTH CALLBACK] writing successful login audit: request_id=%s client_id=%s user_id_present=%v",
		requestID,
		h.config.KeycloakWebClientID,
		userID != "",
	)

	if err := h.auditService.LoginResult(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		true,
		h.config.KeycloakWebClientID,
		c.ClientIP(),
		c.Request.UserAgent(),
		"",
		"",
	); err != nil {
		log.Printf(
			"[AUTH CALLBACK] successful login audit failed: request_id=%s user_id_present=%v err=%v",
			requestID,
			userID != "",
			err,
		)
	}

	log.Printf(
		"[AUTH CALLBACK] setting access token cookie: request_id=%s expires_in=%d",
		requestID,
		tokens.ExpiresIn,
	)

	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)

	if tokens.RefreshToken != "" && tokens.RefreshExpiresIn > 0 {
		log.Printf(
			"[AUTH CALLBACK] setting refresh token cookie: request_id=%s refresh_expires_in=%d",
			requestID,
			tokens.RefreshExpiresIn,
		)

		h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)
	} else {
		log.Printf(
			"[AUTH CALLBACK] refresh token cookie skipped: request_id=%s refresh_token_present=%v refresh_expires_in=%d",
			requestID,
			tokens.RefreshToken != "",
			tokens.RefreshExpiresIn,
		)
	}

	if tokens.IDToken != "" {
		log.Printf(
			"[AUTH CALLBACK] setting id token cookie: request_id=%s expires_in=%d",
			requestID,
			tokens.ExpiresIn,
		)

		h.setSecureIDTokenCookie(c, tokens.IDToken, int(tokens.ExpiresIn))
	} else {
		log.Printf(
			"[AUTH CALLBACK] id token cookie skipped: request_id=%s id_token_present=false",
			requestID,
		)
	}

	// ------------------------------------------------------------------
	// Role-based redirect
	// ------------------------------------------------------------------
	log.Printf(
		"[AUTH CALLBACK] resolving role-based redirect: request_id=%s",
		requestID,
	)

	redirectURL := h.redirectAfterLogin(tokens.AccessToken)

	log.Printf(
		"[AUTH CALLBACK] redirecting after login: request_id=%s redirect_url=%s",
		requestID,
		redirectURL,
	)

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

		// Missing refresh_token is usually normal:
		// user not logged in, initial app load, cookies cleared, or public page.
		// Do not notify admins and do not clear OAuth login-flow cookies.
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

		_ = h.auditService.TokenRefresh(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		// Important:
		// Refresh failure should clear session cookies only.
		// Do not clear oauth_state/pkce_verifier because a login flow may be in progress.
		h.clearSessionCookies(c)

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

		// Same rule: clear session cookies only.
		h.clearSessionCookies(c)

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

	// Logout clears everything.
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

func (h *Handler) redirectAfterLogin(accessToken string) string {
	baseURL := strings.TrimRight(h.config.FrontendBaseURL, "/")

	log.Printf(
		"[AUTH REDIRECT] resolving post-login redirect: frontend_base_url=%q normalized_base_url=%q token_present=%v",
		h.config.FrontendBaseURL,
		baseURL,
		accessToken != "",
	)

	if baseURL == "" {
		log.Printf("[AUTH REDIRECT] empty frontend base url, redirecting to root")
		return "/"
	}

	isAdmin := utils.TokenHasRealmRole(accessToken, "admin")
	isUser := utils.TokenHasRealmRole(accessToken, "user")

	log.Printf(
		"[AUTH REDIRECT] token role checks: is_admin=%v is_user=%v",
		isAdmin,
		isUser,
	)

	if isAdmin {
		redirectURL := baseURL + "/admin/home"

		log.Printf(
			"[AUTH REDIRECT] admin role matched, redirecting to %q",
			redirectURL,
		)

		return redirectURL
	}

	if isUser {
		redirectURL := baseURL + "/apps/news"

		log.Printf(
			"[AUTH REDIRECT] user role matched, redirecting to %q",
			redirectURL,
		)

		return redirectURL
	}

	redirectURL := baseURL + "/"

	log.Printf(
		"[AUTH REDIRECT] no supported role matched, redirecting to default %q",
		redirectURL,
	)

	return redirectURL
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

func (h *Handler) clearSessionCookies(c *gin.Context) {
	h.clearCookie(c, cookieAccessToken, true)
	h.clearCookie(c, cookieRefreshToken, true)
	h.clearCookie(c, cookieIDToken, true)
}

func (h *Handler) clearOAuthCookies(c *gin.Context) {
	h.clearCookie(c, cookiePKCEVerifier, true)
	h.clearCookie(c, cookieOAuthState, true)
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	h.clearSessionCookies(c)
	h.clearOAuthCookies(c)
}

// ----------------------------------------------------
// AUTH FLOW HELPERS
// ----------------------------------------------------
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
