package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv-web/cookie"
	"github.com/ystv/stv-web/structs"
	"golang.org/x/oauth2"
)

const COOKIE string = "session"
const PKCE_COOKIE string = "pkce"
const STATE string = "state"
const UMA_GRANT_TYPE string = "urn:ietf:params:oauth:grant-type:uma-ticket"
const PERMISSION string = "stv-admin#stv-admin"

type (
	LiteToken struct {
		Access  string
		Refresh string
		Expiry  time.Time
	}
	Profile struct {
		Token LiteToken
		Name  string
		Email string
	}
	AuthConf struct {
		Provider oidc.Provider
		Oauth2   oauth2.Config
		Verifier oidc.IDTokenVerifier
	}
)

func (l *LiteToken) MakeToken() oauth2.Token {
	return oauth2.Token{
		AccessToken:  l.Access,
		TokenType:    "Bearer",
		RefreshToken: l.Refresh,
		Expiry:       l.Expiry,
		ExpiresIn:    0,
	}
}

func MakeLiteToken(t *oauth2.Token) LiteToken {
	return LiteToken{
		Access:  t.AccessToken,
		Refresh: t.RefreshToken,
		Expiry:  t.Expiry,
	}
}

var store = map[[32]byte]Profile{}

func (p *Profile) Hash() [32]byte {
	return sha256.Sum256([]byte(p.Name + p.Email + p.Token.Access + p.Token.Refresh))
}

func GetProfile(c echo.Context) (*Profile, error) {
	session, err := cookie.GetSession(c, COOKIE)
	if err != nil {
		return nil, err
	}
	hash, ok := session.Values["profile"].([32]byte)
	profile, ok := store[hash]
	if !ok {
		return nil, errors.New("profile asked for but does not exist!")
	}
	return &profile, nil
}

func AuthMw(kc structs.Keycloak) echo.MiddlewareFunc {
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
			profile, ok := store[hash]
			if session.IsNew || !ok {
				if err = session.Save(c.Request(), c.Response()); err != nil {
					return err
				}
				return oidcRedirect(c, kc)
			}

			// auth_conf, err := OidcConf(kc, c.Request().Context())

			split_url := strings.Split(strings.Trim(c.Request().URL.Path, "/"), "/")
			if split_url[0] == "admin" {
				token := profile.Token.MakeToken()
				permission, err := umaPermissionReq(
					token,
					[]string{"stv-admin#stv-admin"},
					kc,
					c.Request().Context(),
				)
				if err != nil {
					return err
				}
				if !permission {
					return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("User not authenticated: %s", profile.Name))
				}
			}
			return next(c) // accept either on /* or with UMA on /admin/*
		}
	}
}

func umaPermissionReq(token oauth2.Token, permissions []string, kc structs.Keycloak, ctx context.Context) (bool, error) {
	conf, err := OidcConf(kc, ctx)
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
	request.Header.Add("Authorization", "Bearer "+token.AccessToken)

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

func OidcConf(kc structs.Keycloak, c context.Context) (*AuthConf, error) {
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
		RedirectURL:  kc.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	idTokenVerifier := provider.Verifier(
		&oidc.Config{ClientID: kc.ClientID},
	)

	return new(AuthConf{
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

	conf, err := OidcConf(kc, c.Request().Context())

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

		auth_conf, err := OidcConf(kc, req.Context())
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
			Token: MakeLiteToken(oauth2token),
			Email: claims.Email,
			Name:  claims.Name,
		}

		hash := profile.Hash()
		store[hash] = profile

		session.Values["profile"] = hash

		if err = session.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		admin, err := umaPermissionReq(
			*oauth2token,
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
		conf, err := OidcConf(kc, c.Request().Context())
		if err != nil {
			return nil
		}

		logout_split := strings.Split(conf.Oauth2.Endpoint.TokenURL, "/")
		logout_split[len(logout_split)-1] = "logout"
		logout_url := strings.Join(logout_split, "/")
		return c.Redirect(http.StatusFound, logout_url)
	}
}
