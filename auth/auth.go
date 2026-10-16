package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv-web/cookie"
	"github.com/ystv/stv-web/structs"
	"golang.org/x/oauth2"
)

const COOKIE string = "session"
const PKCE_COOKIE string = "pkce"
const STATE string = "state"
const UMA_GRANT_TYPE string = "grant_type=urn:ietf:params:oauth:grant-type:uma-ticket"
const PERMISSION string = "stv-admin#stv-admin"

type (
	Profile struct {
		Client *http.Client
		Name   string
		Email  string
	}
	AuthConf struct {
		Provider oidc.Provider
		Oauth2   oauth2.Config
		Verifier oidc.IDTokenVerifier
	}
)

var Auth *AuthConf = nil

func AuthMw(kc structs.Keycloak) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.Trim(c.Request().URL.Path, "/") == "oath/callback" {
				// if oauth callback, we're making a cookie so ignore this for now
				return next(c)
			}
			session, err := cookie.GetSession(c, COOKIE)
			if err != nil {
				return err
			}

			profile, ok := session.Values["profile"].(Profile)

			if session.IsNew || !ok {
				if err = session.Save(c.Request(), c.Response()); err != nil {
					return err
				}
				return oidcRedirect(c, kc)
			}

			split_url := strings.Split(strings.Trim(c.Request().URL.Path, "/"), "/")
			if split_url[0] == "admin" {
				permission, err := umaPermissionReq(
					profile.Client,
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

func umaPermissionReq(client *http.Client, permissions []string, kc structs.Keycloak, ctx context.Context) (bool, error) {
	uri := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", strings.TrimSuffix(kc.URL, "/"), strings.Trim(kc.Realm, "/"))
	permissions_urlenc := []string{}

	for _, permission := range permissions {
		permissions_urlenc = append(
			permissions_urlenc,
			fmt.Sprintf("permission=%s", permission),
		)
	}
	preamble := []string{
		UMA_GRANT_TYPE,
		fmt.Sprintf("audience=%s", kc.ClientID),
		"response_mode=decision",
	}
	body := strings.Join(
		append(
			preamble,
			permissions_urlenc...,
		),
		"&",
	)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(body))
	if err != nil {
		return false, err
	}

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	if string(response.Status[0]) != "2" {
		return false, errors.New(response.Status)
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
	if Auth != nil {
		return Auth, nil
	}

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

	Auth = new(AuthConf{
		Provider: *provider,
		Oauth2:   oauth2conf,
		Verifier: *idTokenVerifier,
	})

	return Auth, nil
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

	return c.Redirect(http.StatusFound, auth_url)
}

func NewOauthCallback(kc structs.Keycloak) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := cookie.GetSession(c, COOKIE)

		req := c.Request()

		if err != nil {
			return err
		}

		state := session.Values[STATE].(string)
		code_verifier := session.Values[PKCE_COOKIE].(string)

		if req.FormValue("state") != state {
			return err
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
			Client: auth_conf.Oauth2.Client(context.Background(), oauth2token),
			Email:  claims.Email,
			Name:   claims.Name,
		}
		session.Values["profile"] = profile

		admin, err := umaPermissionReq(
			profile.Client,
			[]string{PERMISSION},
			kc,
			req.Context(),
		)

		if err != nil || !admin {
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
