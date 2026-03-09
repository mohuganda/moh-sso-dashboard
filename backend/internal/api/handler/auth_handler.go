package handler

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type AuthHandler struct {
	authService         service.AuthService
	auditService        *service.AuditService
	notificationService service.NotificationsService
	config              *config.Config
}

func NewAuthHandler(
	authService service.AuthService,
	auditService *service.AuditService,
	notificationService service.NotificationsService,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		auditService:        auditService,
		notificationService: notificationService,
		config:              config,
	}
}

// ----------------------------------------------------
// LOGIN (redirect → Keycloak with PKCE)
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthLogin(c *gin.Context) {
	_ = h.auditService.LoginInitiated(
		c.Request.Context(),
		c.ClientIP(),
		c.Request.UserAgent(),
		h.config.KeycloakWebClientID,
	)

	// 🔐 PKCE
	codeVerifier := utils.GenerateCodeVerifier()
	codeChallenge := utils.GenerateCodeChallenge(codeVerifier)

	// Store verifier securely (env-agnostic)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "pkce_verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set true when TLS is enabled
		SameSite: http.SameSiteLaxMode,
	})

	authURL, err := url.Parse(
		h.config.KeycloakExternalURL +
			"/realms/" + h.config.KeycloakRealm +
			"/protocol/openid-connect/auth",
	)
	if err != nil {
		log.Println("failed to parse Keycloak URL:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	q := authURL.Query()
	q.Set("client_id", h.config.KeycloakWebClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid")
	q.Set("redirect_uri", h.config.KeycloakRedirectUri)

	// 🔐 PKCE params
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")

	authURL.RawQuery = q.Encode()
	c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

// ----------------------------------------------------
// GET CURRENT USER (API)
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthGetMe(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
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
// OIDC CALLBACK (PKCE verification)
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		h.auditLoginFailure(c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing authorization code",
		})
		return
	}

	// 🔐 Read PKCE verifier
	codeVerifier, err := c.Cookie("pkce_verifier")
	if err != nil || codeVerifier == "" {
		h.auditLoginFailure(c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "missing pkce verifier",
			"details": err.Error(),
		})
		return
	}

	tokens, err := h.authService.ProcessAuthCode(code, codeVerifier)
	if err != nil {
		h.notificationService.NotifyLoginFailed(
			c.Request.Context(),
			h.config.KeycloakWebClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		h.auditLoginFailure(c)

		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": "authentication failed", "details": err.Error()},
		)
		return
	}

	// Clear PKCE verifier (one-time use)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "pkce_verifier",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

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
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)
	h.setSecureIDTokenCookie(c, tokens.IDToken, int(tokens.ExpiresIn))
	// ------------------------------------------------------------------
	// Role-based redirect (authoritative)
	// ------------------------------------------------------------------

	isAdmin := utils.TokenHasRealmRole(tokens.AccessToken, "admin")
	isUser := utils.TokenHasRealmRole(tokens.AccessToken, "user")

	baseURL := h.config.FrontendBaseURL // e.g. http://localhost:3000

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
func (h *AuthHandler) HandleAuthRefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
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

		h.setSecureRefreshTokenCookie(c, "", -1)

		response.Fail(
			c,
			http.StatusUnauthorized,
			apierror.ErrTokenInvalid.Code,
			"Invalid refresh token",
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
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)

	response.OK(c, http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"expires_in":   tokens.ExpiresIn,
	})
}

// ----------------------------------------------------
// LOGOUT
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthLogout(c *gin.Context) {
	ctx := c.Request.Context()

	userID := utils.ToNullUUID(c.GetString("user_id"))

	// Audit logout
	_ = h.auditService.Logout(
		ctx,
		userID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	// Read tokens from cookies
	refreshToken, _ := c.Cookie("refresh_token")
	idTokenHint, _ := c.Cookie("id_token")

	// Revoke refresh token (server-side logout)
	if refreshToken != "" {
		if err := h.authService.LogOut(
			refreshToken,
		); err != nil {
			log.Printf("[AUTH LOGOUT] token revocation failed: %v", err)
		}
	}

	// Build Keycloak logout redirect URL
	redirectURL := h.authService.GetLogoutURL(
		idTokenHint,
		h.config.FrontendBaseURL,
	)

	// Clear cookies
	clear := func(name string, httpOnly bool) {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: httpOnly,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	}

	clear("refresh_token", true)
	clear("access_token", false)
	clear("id_token", false)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------
func (h *AuthHandler) auditLoginFailure(c *gin.Context) {
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
// COOKIE HELPERS
// ----------------------------------------------------
func (h *AuthHandler) setSecureAccessTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setSecureRefreshTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setSecureIDTokenCookie(c *gin.Context, token string, expiresIn int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "id_token",
		Value:    token,
		Path:     "/",
		MaxAge:   expiresIn,
		HttpOnly: false, // Keycloak logout redirect may require browser visibility
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
