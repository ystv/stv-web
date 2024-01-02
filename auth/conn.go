package auth

import (
	"crypto/tls"
	"errors"
	"fmt"

	ldap "github.com/go-ldap/ldap/v3"
)

// Conn represents an Active Directory connection.
type Conn struct {
	Conn   *ldap.Conn
	Config *Config
}

// Connect returns an open connection to an Active Directory server or an error if one occurred.
func (c *Config) Connect() (*Conn, error) {
	switch c.Security {
	case SecurityNone:
		conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Server, c.Port))
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		return &Conn{Conn: conn, Config: c}, nil
	case SecurityTLS:
		//nolint:gosec
		conn, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.Server, c.Port), &tls.Config{ServerName: c.Server, RootCAs: c.RootCAs})
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		return &Conn{Conn: conn, Config: c}, nil
	case SecurityStartTLS:
		conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Server, c.Port))
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		//nolint:gosec
		err = conn.StartTLS(&tls.Config{ServerName: c.Server, RootCAs: c.RootCAs})
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		return &Conn{Conn: conn, Config: c}, nil
	case SecurityInsecureTLS:
		//nolint:gosec
		conn, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.Server, c.Port), &tls.Config{ServerName: c.Server, InsecureSkipVerify: true})
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		return &Conn{Conn: conn, Config: c}, nil
	case SecurityInsecureStartTLS:
		conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Server, c.Port))
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		//nolint:gosec
		err = conn.StartTLS(&tls.Config{ServerName: c.Server, InsecureSkipVerify: true})
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		return &Conn{Conn: conn, Config: c}, nil
	default:
		return nil, errors.New("configuration error: invalid SecurityType")
	}
}

// Bind authenticates the connection with the given userPrincipalName and password
// and returns the result or an error if one occurred.
func (c *Conn) Bind(upn, password string) (bool, error) {
	if password == "" {
		return false, nil
	}

	err := c.Conn.Bind(upn, password)
	if err != nil {
		var e *ldap.Error
		if errors.As(err, &e) {
			if e.ResultCode == ldap.LDAPResultInvalidCredentials {
				return false, nil
			}
		}
		return false, fmt.Errorf("bind error (%s): %w", upn, err)
	}

	return true, nil
}
