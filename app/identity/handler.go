package identity

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/WeCanHearYou/wechy/app"
	"github.com/labstack/echo"
)

type oauthUserProfile struct {
	Name  string
	ID    string
	Email string
}

// OAuthHandlers contains multiple oauth HTTP handlers
type OAuthHandlers struct {
	tenantService TenantService
	oauthService  OAuthService
	userService   UserService
}

// OAuth creates a new OAuthHandlers
func OAuth(tenantService TenantService, oauthService OAuthService, userService UserService) OAuthHandlers {
	return OAuthHandlers{tenantService, oauthService, userService}
}

// Callback handles OAuth callbacks
func (h OAuthHandlers) Callback(provider string) app.HandlerFunc {
	return func(c app.Context) error {

		redirect := c.QueryParam("state")
		code := c.QueryParam("code")
		redirectURL, err := url.Parse(redirect)

		if err != nil {
			c.Logger().Errorf("Could not parse url %s", redirect)
			return c.Render(http.StatusInternalServerError, "500.html", echo.Map{})
		}

		tenant, err := h.tenantService.GetByDomain(stripPort(redirectURL.Host))
		if err != nil {
			return err
		}

		//TODO: Check if code is empty (or other querystring parameter)
		//Because the user can deny access to it
		oauthUser, err := h.oauthService.GetProfile(provider, code)
		if err != nil {
			c.Logger().Error(err)
			return c.Render(http.StatusInternalServerError, "500.html", echo.Map{})
		}

		user, err := h.userService.GetByEmail(oauthUser.Email)
		if err != nil {
			if err == app.ErrNotFound {
				user = &app.User{
					Name:   oauthUser.Name,
					Tenant: tenant,
					Email:  oauthUser.Email,
					Providers: []*app.UserProvider{
						&app.UserProvider{
							UID:  oauthUser.ID,
							Name: provider,
						},
					},
				}

				err = h.userService.Register(user)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		claims := &app.WechyClaims{
			UserID:    user.ID,
			UserName:  user.Name,
			UserEmail: user.Email,
			TenantID:  tenant.ID,
		}

		var token string
		if token, err = Encode(claims); err != nil {
			c.Logger().Errorf("Encoding claims failed with %s", err)
			return c.Render(http.StatusInternalServerError, "500.html", echo.Map{})
		}

		var query = redirectURL.Query()
		query.Add("jwt", token)
		redirectURL.RawQuery = query.Encode()
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
	}
}

// Login handles OAuth logins
func (h OAuthHandlers) Login(provider string) app.HandlerFunc {
	return func(c app.Context) error {
		authURL := h.oauthService.GetAuthURL(provider, c.QueryParam("redirect"))
		return c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// Logout remove auth cookies
func (h OAuthHandlers) Logout() app.HandlerFunc {
	return func(c app.Context) error {
		c.SetCookie(&http.Cookie{
			Name:    "auth",
			MaxAge:  -1,
			Expires: time.Now().Add(-100 * time.Hour),
		})
		return c.Redirect(http.StatusTemporaryRedirect, c.QueryParam("redirect"))
	}
}

func stripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	return hostport[:colon]
}