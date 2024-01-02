package auth

import (
	"github.com/go-ldap/ldap/v3"
)

// Authenticate checks if the given credentials are valid, or returns an error if one occurred.
// Username may be either the sAMAccountName or the userPrincipalName.
func Authenticate(config Config, username, password string) (bool, error) {
	upn, err := config.UPN(username)
	if err != nil {
		return false, err
	}

	conn, err := config.Connect()
	if err != nil {
		return false, err
	}
	defer func(Conn *ldap.Conn) {
		_ = Conn.Close()
	}(conn.Conn)

	return conn.Bind(upn, password)
}

// AuthenticateExtended checks if the given credentials are valid, or returns an error if one occurred.
// Username may be either the sAMAccountName or the userPrincipalName.
// Entry is the *ldap.Entry that holds the DN and any request attributes of the user.
// If groups are non-empty, userGroups will hold which of those groups the user is a member of.
// Groups can be a list of groups referenced by DN or cn and the format provided will be the format returned.
func AuthenticateExtended(config Config, username, password string, attrs, groups []string) (status bool, entry *ldap.Entry, userGroups []string, err error) {
	upn, err := config.UPN(username)
	if err != nil {
		return false, nil, nil, err
	}

	conn, err := config.Connect()
	if err != nil {
		return false, nil, nil, err
	}
	defer func(Conn *ldap.Conn) {
		_ = Conn.Close()
	}(conn.Conn)

	// bind
	status, err = conn.Bind(upn, password)
	if err != nil {
		return false, nil, nil, err
	}
	if !status {
		return false, nil, nil, nil
	}

	// get entry
	entry, err = conn.GetAttributes("userPrincipalName", upn, attrs)
	if err != nil {
		return false, nil, nil, err
	}

	if len(groups) > 0 {
		// get all groups
		var foundGroups []*ldap.Entry
		foundGroups, err = conn.getGroups(entry.DN)
		if err != nil {
			return false, nil, nil, err
		}

		for _, group := range groups {
			var groupDN string
			groupDN, err = conn.GroupDN(group)
			if err != nil {
				return false, nil, nil, err
			}

			for _, userGroup := range foundGroups {
				if userGroup.DN == groupDN {
					userGroups = append(userGroups, group)
					break
				}
			}
		}
	}

	return status, entry, userGroups, nil
}
