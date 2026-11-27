package client

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type UserInfo struct {
	Username  string   `json:"username"`
	SuperUser bool     `json:"superUser"`
	Protected bool     `json:"protected"`
	Databases []string `json:"databases"`
}

type UserClient interface {
	ExecSQL(command string, timeout uint) error
	ExecSQLForRows(command string, timeout uint) ([]string, error)
}

func ListUsers(cli UserClient) ([]UserInfo, error) {
	rows, err := cli.ExecSQLForRows(`SELECT COALESCE(json_agg(u ORDER BY u.username), '[]'::json)::text FROM (
        SELECT r.rolname AS username, r.rolsuper AS "superUser",
            (r.rolname = current_user OR r.rolname = 'postgres' OR r.rolname LIKE 'pg\_%') AS protected,
            COALESCE((SELECT json_agg(d.datname ORDER BY d.datname) FROM pg_database d
                WHERE NOT d.datistemplate AND EXISTS (
                    SELECT 1 FROM aclexplode(d.datacl) a WHERE a.grantee = r.oid
                )), '[]'::json) AS databases
        FROM pg_roles r WHERE r.rolcanlogin
    ) u`, 300)
	if err != nil {
		return nil, err
	}
	var users []UserInfo
	err = json.Unmarshal([]byte(strings.TrimSpace(strings.Join(rows, "\n"))), &users)
	return users, err
}

func GrantUser(cli UserClient, database, username string, revoke bool) error {
	statement := "GRANT ALL PRIVILEGES ON DATABASE " + QuoteIdentifier(database) + " TO " + QuoteIdentifier(username)
	if revoke {
		statement = "REVOKE ALL PRIVILEGES ON DATABASE " + QuoteIdentifier(database) + " FROM " + QuoteIdentifier(username)
	}
	return cli.ExecSQL(statement, 300)
}

func QuoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func QuoteLiteral(value string) string {
	return `E'` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `''`) + `'`
}

func (r *Remote) ExecSQLForRows(command string, timeout uint) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rows, err := r.Client.QueryContext(ctx, command)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
