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
	"github.com/ystv/stv-web/store"
	"github.com/ystv/stv-web/structs"
	"golang.org/x/oauth2"
)

const COOKIE string = "session"
const PKCE_COOKIE string = "pkce"
const STATE string = "state"
const UMA_GRANT_TYPE string = "urn:ietf:params:oauth:grant-type:uma-ticket"
const PERMISSION string = "stv-admin#stv-admin"

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

var local_store *store.Store

func (p *Profile) Hash() [32]byte {
	return sha256.Sum256([]byte(p.Name + p.Email))
}

func GetProfile(c echo.Context) (*Profile, error) {
	session, err := cookie.GetSession(c, COOKIE)
	if err != nil {
		return nil, err
	}
	hash, ok := session.Values["profile"].([32]byte)
	if !ok {
		return nil, fmt.Errorf("Profile lost by client!")
	}
	s, err := local_store.GetSessionHash(hash[:], nil)
	if err != nil {
		return nil, err
	}
	profile := ProfileFromSession(s)
	return &profile, nil
}

func AuthMw(_store *store.Store, kc structs.Keycloak) echo.MiddlewareFunc {
	local_store = _store
	gob.Register([32]byte{})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/oauth/callback" {
				// if oauth callback, we're making a cookie so ignore this for now
				return next(c)
			}
			session, err := cookie.GetSession(c, COOKIE)
			if err != nil && session == nil {
				return err
			}
			hash, ok := session.Values["profile"].([32]byte)
			s, err := local_store.GetSessionHash(hash[:], nil)
			if session.IsNew || !ok || s == nil {
				if err = session.Save(c.Request(), c.Response()); err != nil {
					return err
				}
				return oidcRedirect(c, kc)
			}
			profile := ProfileFromSession(s)

			split_url := strings.Split(strings.Trim(c.Request().URL.Path, "/"), "/")
			token := profile.Token
			var requested_permission []string
			if split_url[0] == "admin" {
				requested_permission = []string{"stv-admin#stv-admin"}
			} else {
				requested_permission = []string{"Default Resource"}
			}
			permission, err := umaPermissionReq(
				token,
				requested_permission,
				kc,
				c.Request().Context(),
			)
			if err != nil {
				local_store.DeleteSession(hash[:], nil)
				return err
			}
			if !permission {
				local_store.DeleteSession(hash[:], nil)
				return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("User not authenticated: %s", profile.Name))
			}
			return next(c) // accept either on /* or with UMA on /admin/*
		}
	}
}

func umaPermissionReq(token string, permissions []string, kc structs.Keycloak, ctx context.Context) (bool, error) {
	conf, err := MakeOidcConf(kc, ctx)
	if err != nil {
		return false, err
	}

	uri := conf.Provider.Endpoint().TokenURL
	form := url.Values{}
	form.Set("grant_type", UMA_GRANT_TYPE)
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

	var resp_body struct {
		Result bool `json:"result"`
	}

	if err = decoder.Decode(&resp_body); err != nil {
		return false, err
	}

	return resp_body.Result, nil
}

func MakeOidcConf(kc structs.Keycloak, c context.Context) (*OidcConf, error) {
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
	session, err := cookie.GetSession(c, COOKIE)
	if err != nil {
		return err
	}

	conf, err := MakeOidcConf(kc, c.Request().Context())

	if err != nil {
		return err
	}

	state := rand.Text()
	code_verifier := oauth2.GenerateVerifier()

	auth_url := conf.Oauth2.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.S256ChallengeOption(code_verifier),
	)
	session.Values[PKCE_COOKIE] = code_verifier
	session.Values[STATE] = state

	if err = session.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, auth_url)
}

func NewOauthCallback(kc structs.Keycloak) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := cookie.GetSession(c, COOKIE)

		if err != nil {
			return err
		}
		state := session.Values[STATE].(string)
		code_verifier := session.Values[PKCE_COOKIE].(string)

		req := c.Request()
		if req.FormValue("state") != state {
			return echo.NewHTTPError(http.StatusUnauthorized, "State did not match state given to oauth")
		}

		auth_conf, err := MakeOidcConf(kc, req.Context())
		if err != nil {
			return err
		}

		oauth2token, err := auth_conf.Oauth2.Exchange(
			req.Context(),
			req.FormValue("code"),
			oauth2.VerifierOption(code_verifier),
		)

		if err != nil {
			return err
		}
		rawIDToken, ok := oauth2token.Extra("id_token").(string)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "ID token inaccessible")
		}

		idToken, err := auth_conf.Verifier.Verify(req.Context(), rawIDToken)
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
		local_store.AddSession(hash[:], profile.Name, profile.Email, profile.Token, nil)

		session.Values["profile"] = hash

		if err = session.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		admin, err := umaPermissionReq(
			profile.Token,
			[]string{PERMISSION},
			kc,
			req.Context(),
		)
		if err != nil {
			return err
		}

		if !admin {
			return c.Redirect(http.StatusFound, "/")
		} else {
			return c.Redirect(http.StatusFound, "/admin")
		}
	}
}

func NewOauthLogout(kc structs.Keycloak) echo.HandlerFunc {
	return func(c echo.Context) error {
		conf, err := MakeOidcConf(kc, c.Request().Context())
		if err != nil {
			return nil
		}

		// if we can't access to delete, then it doesn't matter... (hopefully)
		session, err := cookie.GetSession(c, COOKIE)
		if err == nil {
			hash, ok := session.Values["profile"].([32]byte)
			if ok {
				local_store.DeleteSession(hash[:], nil)
			}
		}

		logout_split := strings.Split(conf.Oauth2.Endpoint.TokenURL, "/")
		logout_split[len(logout_split)-1] = "logout"
		logout_url := strings.Join(logout_split, "/")
		return c.Redirect(http.StatusFound, logout_url)
	}
}
