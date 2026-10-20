package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv-web/cookie"
	"github.com/ystv/stv-web/storage"
	_store "github.com/ystv/stv-web/store"
	"github.com/ystv/stv-web/structs"
	"golang.org/x/oauth2"
)

const Cookie string = "session"
const PkceCookie string = "pkce"
const State string = "state"
const UmaGrantType string = "urn:ietf:params:oauth:grant-type:uma-ticket"
const Permission string = "stv-admin#stv-admin"

type (
	Profile struct {
		Token string
		Name  string
		Email string
	}
	OidcConf struct {
		Provider oidc.Provider
		Oauth2   oauth2.Config
		Verifier oidc.IDTokenVerifier
	}
)

func MakeToken(token string) oauth2.Token {
	return oauth2.Token{
		AccessToken: token,
	}
}

func ProfileFromSession(session *storage.Session) Profile {
	return Profile{
		Token: session.GetAccess(),
		Name:  session.GetName(),
		Email: session.GetEmail(),
	}
}

var store *_store.Store

func (p *Profile) Hash() [32]byte {
	return sha256.Sum256([]byte(p.Name + p.Email))
}

func GetProfile(c echo.Context) (*Profile, error) {
	session, err := cookie.GetSession(c, Cookie)
	if err != nil {
		return nil, err
	}
	hash, ok := session.Values["profile"].([32]byte)
	if !ok {
		return nil, fmt.Errorf("profile lost by client")
	}
	s, err := store.GetSessionHash(hash[:], nil)
	if err != nil {
		return nil, err
	}
	profile := ProfileFromSession(s)
	return &profile, nil
}

func Middleware(s *_store.Store, kc structs.Keycloak) echo.MiddlewareFunc {
	store = s
	gob.Register([32]byte{})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/oauth/callback" {
				// if oauth callback, we're making a cookie so ignore this for now
				return next(c)
			}
			session, err := cookie.GetSession(c, Cookie)
			if err != nil && session == nil {
				return err
			}
			hash, ok := session.Values["profile"].([32]byte)
			s, err := store.GetSessionHash(hash[:], nil)
			if session.IsNew || !ok || s == nil {
				return oidcRedirect(c, kc)
			}
			if err != nil {
				return err
			}

			profile := ProfileFromSession(s)

			splitURL := strings.Split(strings.Trim(c.Request().URL.Path, "/"), "/")
			token := profile.Token
			var requestedPermission []string
			if splitURL[0] == "admin" {
				requestedPermission = []string{"stv-admin#stv-admin"}
			} else {
				requestedPermission = []string{"Default Resource"}
			}
			permission, err := umaPermissionReq(
				c.Request().Context(),
				token,
				requestedPermission,
				kc,
			)
			if err != nil {
				//nolint:errcheck // This only fails if no cookie exists but we're trying to get rid of it anyway...
				_ = store.DeleteSession(hash[:], nil)
				return err
			}
			if !permission {
				//nolint:errcheck // This only fails if no cookie exists but we're trying to get rid of it anyway...
				_ = store.DeleteSession(hash[:], nil)
				return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("user not authenticated: %s", profile.Name))
			}
			return next(c) // accept either on /* or with UMA on /admin/*
		}
	}
}

func umaPermissionReq(ctx context.Context, token string, permissions []string, kc structs.Keycloak) (bool, error) {
	conf, err := MakeOidcConf(ctx, kc)
	if err != nil {
		return false, err
	}

	uri := conf.Provider.Endpoint().TokenURL
	form := url.Values{}
	form.Set("grant_type", UmaGrantType)
	form.Set("audience", kc.ClientID)
	form.Set("response_mode", "decision")
	for _, permission := range permissions {
		form.Add("permission", permission)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Bearer "+token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		body, _ := io.ReadAll(response.Body)
		return false, fmt.Errorf("%s: %s", response.Status, body)
	}

	decoder := json.NewDecoder(response.Body)

	var respBody struct {
		Result bool `json:"result"`
	}

	if err = decoder.Decode(&respBody); err != nil {
		return false, err
	}

	return respBody.Result, nil
}

func MakeOidcConf(c context.Context, kc structs.Keycloak) (*OidcConf, error) {
	url := fmt.Sprintf("%s/realms/%s", strings.TrimSuffix(kc.URL, "/"), strings.Trim(kc.Realm, "/"))
	provider, err := oidc.NewProvider(
		c,
		url,
	)
	if err != nil {
		return nil, err
	}

	oauth2conf := oauth2.Config{
		ClientID:     kc.ClientID,
		ClientSecret: kc.Secret,
		RedirectURL:  kc.Address + "/oauth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	idTokenVerifier := provider.Verifier(
		&oidc.Config{ClientID: kc.ClientID},
	)

	return new(OidcConf{
		Provider: *provider,
		Oauth2:   oauth2conf,
		Verifier: *idTokenVerifier,
	}), nil
}

func oidcRedirect(c echo.Context, kc structs.Keycloak) error {
	session, err := cookie.GetSession(c, Cookie)
	if err != nil && session == nil {
		return err
	}

	conf, err := MakeOidcConf(c.Request().Context(), kc)

	if err != nil {
		return err
	}

	state := rand.Text()
	codeVerifier := oauth2.GenerateVerifier()

	authURL := conf.Oauth2.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.S256ChallengeOption(codeVerifier),
	)

	session.Values[PkceCookie] = codeVerifier
	session.Values[State] = state

	if err = session.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, authURL)
}

func NewOauthCallback(kc structs.Keycloak) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := cookie.GetSession(c, Cookie)

		if err != nil {
			return err
		}
		state := session.Values[State].(string)
		codeVerifier := session.Values[PkceCookie].(string)

		req := c.Request()
		if req.FormValue("state") != state {
			return echo.NewHTTPError(http.StatusUnauthorized, "State did not match state given to oauth")
		}

		authConf, err := MakeOidcConf(req.Context(), kc)
		if err != nil {
			return err
		}

		oauth2token, err := authConf.Oauth2.Exchange(
			req.Context(),
			req.FormValue("code"),
			oauth2.VerifierOption(codeVerifier),
		)

		if err != nil {
			return err
		}
		rawIDToken, ok := oauth2token.Extra("id_token").(string)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "ID token inaccessible")
		}

		idToken, err := authConf.Verifier.Verify(req.Context(), rawIDToken)
		if err != nil {
			return err
		}
		var claims struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		if err := idToken.Claims(&claims); err != nil {
			return err
		}

		profile := Profile{
			Token: oauth2token.AccessToken,
			Email: claims.Email,
			Name:  claims.Name,
		}

		hash := profile.Hash()
		_, err = store.AddSession(hash[:], profile.Name, profile.Email, profile.Token, nil)
		if err != nil {
			return err
		}

		session.Values["profile"] = hash

		if err = session.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		admin, err := umaPermissionReq(
			req.Context(),
			profile.Token,
			[]string{Permission},
			kc,
		)
		if err != nil {
			return err
		}

		if admin {
			return c.Redirect(http.StatusFound, "/admin")
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func NewOauthLogout(kc structs.Keycloak) echo.HandlerFunc {
	return func(c echo.Context) error {
		conf, err := MakeOidcConf(c.Request().Context(), kc)
		if err != nil {
			return err
		}

		// if we can't access to delete, then it doesn't matter... (hopefully)
		session, err := cookie.GetSession(c, Cookie)
		if err == nil {
			hash, ok := session.Values["profile"].([32]byte)
			if ok {
				//nolint:errcheck // Again, if it doesn't exist, we don't care!
				store.DeleteSession(hash[:], nil)
			}
		}

		logoutSplit := strings.Split(conf.Oauth2.Endpoint.TokenURL, "/")
		logoutSplit[len(logoutSplit)-1] = "logout"
		logoutURL := strings.Join(logoutSplit, "/")
		return c.Redirect(http.StatusFound, logoutURL)
	}
}
