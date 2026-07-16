package client

import (
	"fmt"
	"strings"
)

func splitHosts(permission string) []string {
	hosts := strings.Split(permission, ",")
	res := make([]string, 0, len(hosts))
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		res = append(res, host)
	}
	if len(res) == 0 {
		return []string{"%"}
	}
	return res
}

func userIdentity(username, host string) string {
	return fmt.Sprintf("'%s'@'%s'", strings.ReplaceAll(username, "'", "''"), strings.ReplaceAll(host, "'", "''"))
}

func userIdentities(username, permission string) []string {
	hosts := splitHosts(permission)
	users := make([]string, 0, len(hosts))
	for _, host := range hosts {
		users = append(users, userIdentity(username, host))
	}
	return users
}

func isUserExistsErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "error 1396")
}

func createUserSQL(user, password string) string {
	return fmt.Sprintf("create user %s identified by '%s';", user, password)
}

func databaseScope(database string) string {
	if database == "*" {
		return "*.*"
	}
	return fmt.Sprintf("`%s`.*", database)
}

func grantDatabaseSQL(database, user string) string {
	return fmt.Sprintf("grant all privileges on %s to %s", databaseScope(database), user)
}

func createUserGrantSQL(info CreateInfo, user string) string {
	grantStr := grantDatabaseSQL(info.Name, user)
	if strings.HasPrefix(info.Version, "5.7") || strings.HasPrefix(info.Version, "5.6") {
		return fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, info.Password)
	}
	return grantStr + " with grant option;"
}

func grantUserSQL(info GrantInfo) string {
	return grantDatabaseSQL(info.Database, userIdentity(info.Username, info.Host)) + " with grant option;"
}

func revokeGrantSQL(info GrantInfo) string {
	return fmt.Sprintf("revoke all privileges on %s from %s;", databaseScope(info.Database), userIdentity(info.Username, info.Host))
}

func revokeGrantOptionSQL(info GrantInfo) string {
	return fmt.Sprintf("revoke grant option on %s from %s;", databaseScope(info.Database), userIdentity(info.Username, info.Host))
}

func dropUserSQL(info UserInfo, version string) string {
	user := userIdentity(info.Username, info.Host)
	if strings.HasPrefix(version, "5.6") {
		return fmt.Sprintf("drop user %s", user)
	}
	return fmt.Sprintf("drop user if exists %s", user)
}

func renameUserSQL(info UserUpdateInfo) string {
	return fmt.Sprintf("rename user %s to %s", userIdentity(info.Username, info.Host), userIdentity(info.Username, info.NewHost))
}

func dropDatabaseSQL(name string) string {
	return fmt.Sprintf("drop database if exists `%s`", name)
}

func changeUserPasswordSQL(user, password, version string) string {
	if strings.HasPrefix(version, "5.7") || strings.HasPrefix(version, "5.6") {
		return fmt.Sprintf("set password for %s = password('%s')", user, password)
	}
	return fmt.Sprintf("ALTER USER %s IDENTIFIED BY '%s';", user, password)
}
