package service

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/google/uuid"
)

func loadClusterPasswordMember(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, saved string) (string, ClusterPasswordMemberResponse, error) {
	repair := req.CurrentPassword == req.NewPassword
	healthy := req.Action == "check" && !repair
	password := saved
	if req.Role == "master" && req.CurrentPassword != "" {
		password = req.CurrentPassword
	}
	if req.Role == "master" && req.Action == "sync" {
		password = req.NewPassword
	}
	if req.Role == "slave" && req.App != constant.AppRedisCluster && req.Action == "sync" {
		result, err := waitClusterPasswordMember(ctx, db, req, req.NewPassword)
		if err == nil {
			err = verifyClusterPasswordAuthentication(ctx, db, req.App)
		}
		return req.NewPassword, result, err
	}
	if req.Role == "master" && req.Action == "rotate" {
		if result, err := probeClusterPasswordMember(ctx, db, req, req.NewPassword, healthy, false); err == nil {
			password = req.NewPassword
			return password, result, verifyClusterPasswordAuthentication(ctx, db, req.App)
		}
	}
	verifyRedisSource := req.Role == "slave" && req.App == constant.AppRedisCluster
	result, err := probeClusterPasswordMember(ctx, db, req, password, healthy, verifyRedisSource)
	if err != nil && repair && req.Role == "slave" && req.App != constant.AppRedisCluster {
		password = req.NewPassword
		result, err = probeClusterPasswordMember(ctx, db, req, password, healthy, false)
	}
	if err != nil {
		return "", result, err
	}
	return password, result, verifyClusterPasswordAuthentication(ctx, db, req.App)
}

func verifyClusterPasswordAuthentication(ctx context.Context, db clusterDatabase, app string) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	random := uuid.NewString()
	switch app {
	case constant.AppMysqlCluster:
		if result, err := mysqlClusterPasswordQuery(ctx, db, random, "SELECT 1"); err == nil && result == "1" {
			return buserr.New("ErrClusterPasswordAuthenticationNotEnforced")
		}
	case constant.AppPostgresqlCluster:
		if result, err := postgresClusterPasswordQuery(ctx, db, random, "SELECT 1"); err == nil && result == "1" {
			return buserr.New("ErrClusterPasswordAuthenticationNotEnforced")
		}
	case constant.AppRedisCluster:
		if result, err := redisClusterPasswordQuery(ctx, db, random, "PING"); err == nil && result == "PONG" {
			return buserr.New("ErrClusterPasswordAuthenticationNotEnforced")
		}
	}
	return ctx.Err()
}

func waitClusterPasswordMember(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string) (ClusterPasswordMemberResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 75*time.Second)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		result, err := probeClusterPasswordMember(ctx, db, req, password, true, false)
		if err == nil {
			return result, nil
		}
		select {
		case <-ctx.Done():
			return result, err
		case <-ticker.C:
		}
	}
}

func probeClusterPasswordMember(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string, healthy, verifyRedisSource bool) (ClusterPasswordMemberResponse, error) {
	var result ClusterPasswordMemberResponse
	if !validClusterPassword(password) {
		return result, buserr.New("ErrClusterPasswordAuthentication")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	var err error
	switch req.App {
	case constant.AppMysqlCluster:
		result, err = probeMySQLClusterPassword(ctx, db, req, password, healthy)
	case constant.AppPostgresqlCluster:
		result, err = probePostgresClusterPassword(ctx, db, req, password, healthy)
	case constant.AppRedisCluster:
		result, err = probeRedisClusterPassword(ctx, db, req, password, healthy, verifyRedisSource)
	}
	if err != nil {
		return result, err
	}
	if result.Identity == "" || req.SourceID != "" && result.Identity != req.SourceID {
		return result, buserr.New("ErrClusterPasswordSourceMismatch")
	}
	return result, nil
}

func probeMySQLClusterPassword(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string, healthy bool) (ClusterPasswordMemberResponse, error) {
	var result ClusterPasswordMemberResponse
	hosts, err := mysqlClusterPasswordQuery(ctx, db, password, "SELECT Host FROM mysql.user WHERE User = 'root' AND Host IN ('localhost', '%') ORDER BY Host")
	if err != nil {
		return result, buserr.New("ErrClusterPasswordAuthentication")
	}
	result.RootHosts = strings.Fields(hosts)
	if len(result.RootHosts) == 0 {
		return result, buserr.New("ErrClusterPasswordMySQLAccounts")
	}
	if req.Role == "master" {
		status, err := mysqlClusterPasswordQuery(ctx, db, password, "SELECT @@server_uuid, @@read_only, @@super_read_only, @@log_bin, @@session.sql_log_bin")
		if err != nil {
			return result, buserr.New("ErrClusterPasswordAuthentication")
		}
		fields := strings.Fields(status)
		if len(fields) != 5 || fields[1] != "0" || fields[2] != "0" || fields[3] != "1" || fields[4] != "1" {
			return result, buserr.New("ErrClusterPasswordMySQLNotWritable")
		}
		result.Identity = fields[0]
		return result, nil
	}
	status, legacy, err := mysqlClusterPasswordReplicaStatus(ctx, db, req.Template, password)
	if err != nil || strings.Count(status, ". row ") != 1 {
		return result, buserr.New("ErrClusterPasswordSourceMismatch")
	}
	fields := clusterPasswordStatusFields(status)
	identity, ioState, sqlState := fields["Source_UUID"], fields["Replica_IO_Running"], fields["Replica_SQL_Running"]
	if legacy {
		identity, ioState, sqlState = fields["Master_UUID"], fields["Slave_IO_Running"], fields["Slave_SQL_Running"]
	}
	if healthy && (ioState != "Yes" || sqlState != "Yes") {
		return result, buserr.New("ErrClusterPasswordReplicationUnhealthy")
	}
	result.Identity = identity
	return result, nil
}

func mysqlClusterPasswordReplicaStatus(ctx context.Context, db clusterDatabase, template, password string) (string, bool, error) {
	query := func(statement string) (string, error) {
		return db.result(ctx, "MYSQL_PWD", password, "mysql", "--host=127.0.0.1", "--user=root",
			"--batch", "--vertical", "--column-names", "--execute", statement)
	}
	if !strings.HasPrefix(template, "5.7") {
		status, err := query("SHOW REPLICA STATUS")
		if err == nil || !strings.HasPrefix(template, "8.0-") && !strings.HasPrefix(template, "8.0.") {
			return status, false, err
		}
	}
	status, err := query("SHOW SLAVE STATUS")
	return status, true, err
}

func mysqlClusterPasswordQuery(ctx context.Context, db clusterDatabase, password, query string) (string, error) {
	return db.result(ctx, "MYSQL_PWD", password, "mysql", "--host=127.0.0.1", "--user=root",
		"--batch", "--skip-column-names", "--execute", query)
}

func probePostgresClusterPassword(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string, healthy bool) (ClusterPasswordMemberResponse, error) {
	var result ClusterPasswordMemberResponse
	status, err := postgresClusterPasswordQuery(ctx, db, password, "SELECT system_identifier, pg_is_in_recovery(), EXISTS (SELECT 1 FROM pg_stat_wal_receiver WHERE status = 'streaming') FROM pg_control_system()")
	if err != nil {
		return result, buserr.New("ErrClusterPasswordAuthentication")
	}
	fields := strings.Split(status, "|")
	if len(fields) != 3 || (fields[1] == "t") != (req.Role == "slave") {
		return result, buserr.New("ErrClusterPasswordSourceMismatch")
	}
	if req.Role == "slave" && healthy && fields[2] != "t" {
		return result, buserr.New("ErrClusterPasswordReplicationUnhealthy")
	}
	result.Identity = fields[0]
	return result, nil
}

func postgresClusterPasswordQuery(ctx context.Context, db clusterDatabase, password, query string) (string, error) {
	return db.result(ctx, "PGPASSWORD", password, "psql", "--no-password", "--host=127.0.0.1",
		"--username=postgres", "--dbname=postgres", "--no-psqlrc", "--tuples-only", "--no-align", "--command", query)
}

func redisClusterPasswordQuery(ctx context.Context, db clusterDatabase, password string, args ...string) (string, error) {
	return db.result(ctx, "REDISCLI_AUTH", password, append([]string{"redis-cli", "--raw"}, args...)...)
}

func probeRedisClusterPassword(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string, healthy, verifySource bool) (ClusterPasswordMemberResponse, error) {
	var result ClusterPasswordMemberResponse
	if req.Role == "master" {
		if ping, err := redisClusterPasswordQuery(ctx, db, password, "PING"); err != nil || ping != "PONG" {
			return result, buserr.New("ErrClusterPasswordAuthentication")
		}
	}
	status, err := redisClusterPasswordQuery(ctx, db, password, "INFO", "replication")
	fields := clusterPasswordStatusFields(status)
	if err != nil || fields["role"] == "" {
		return result, buserr.New("ErrClusterPasswordAuthentication")
	}
	if fields["role"] != req.Role {
		return result, buserr.New("ErrClusterPasswordSourceMismatch")
	}
	if req.Role == "slave" {
		host, port := fields["master_host"], fields["master_port"]
		if host == "" || host != db.env["MASTER_HOST"] || port != db.env["MASTER_PORT"] {
			return result, buserr.New("ErrClusterPasswordSourceMismatch")
		}
		if healthy && (fields["master_link_status"] != "up" || fields["master_sync_in_progress"] != "0") {
			return result, buserr.New("ErrClusterPasswordReplicationUnhealthy")
		}
		if verifySource {
			sourcePassword := req.NewPassword
			if healthy {
				if fields["master_replid"] != req.SourceID {
					return result, buserr.New("ErrClusterPasswordSourceMismatch")
				}
				sourcePassword = db.env[clusterSourcePassword]
			}
			identity, err := probeRedisClusterPasswordSource(ctx, db, host, port, sourcePassword)
			if err != nil {
				return result, err
			}
			result.Identity = identity
			return result, nil
		}
	}
	result.Identity = fields["master_replid"]
	return result, nil
}

func probeRedisClusterPasswordSource(ctx context.Context, db clusterDatabase, host, port, password string) (string, error) {
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 || strings.HasPrefix(host, "-") {
		return "", buserr.New("ErrClusterPasswordSourceMismatch")
	}
	if !validClusterPassword(password) {
		return "", buserr.New("ErrClusterPasswordReplicaSourceAuthentication")
	}
	output, err := db.command(ctx, "REDISCLI_AUTH", password, "redis-cli", "--raw", "-h", host, "-p", port, "INFO", "replication").Output()
	response := strings.TrimSpace(string(output))
	var exitErr *exec.ExitError
	if strings.HasPrefix(response, "NOAUTH ") || strings.HasPrefix(response, "WRONGPASS ") || strings.HasPrefix(response, "NOPERM ") ||
		errors.As(err, &exitErr) && strings.HasPrefix(strings.TrimSpace(string(exitErr.Stderr)), "AUTH failed:") {
		return "", buserr.New("ErrClusterPasswordReplicaSourceAuthentication")
	}
	source := clusterPasswordStatusFields(response)
	if err != nil || source["role"] == "" {
		return "", buserr.New("ErrClusterPasswordReplicaSourceUnavailable")
	}
	if source["role"] != "master" || source["master_replid"] == "" {
		return "", buserr.New("ErrClusterPasswordSourceMismatch")
	}
	return source["master_replid"], nil
}

func clusterPasswordStatusFields(result string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(result, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if ok {
			fields[key] = strings.TrimSpace(value)
		}
	}
	return fields
}

func rotateClusterSQLPassword(ctx context.Context, db clusterDatabase, req ClusterPasswordMemberRequest, password string, hosts []string) error {
	if req.App == constant.AppPostgresqlCluster {
		input := strings.NewReader(req.NewPassword + "\n" + req.NewPassword + "\n")
		if err := db.exec(ctx, io.Discard, input, "PGPASSWORD", password, "psql", "--no-psqlrc",
			"--no-password", "--host=127.0.0.1", "--username=postgres", "--dbname=postgres", "--set=ON_ERROR_STOP=1",
			"--command", "\\password"); err != nil {
			return clusterPasswordCommandError(ctx)
		}
		return nil
	}
	quoted := "'" + strings.ReplaceAll(req.NewPassword, "'", "''") + "'"
	accounts, err := mysqlClusterPasswordQuery(ctx, db, password,
		"SELECT Host, plugin FROM mysql.user WHERE User = 'root' AND Host IN ('localhost', '%') ORDER BY Host")
	if err != nil {
		return buserr.New("ErrClusterPasswordMySQLAccounts")
	}
	lines := strings.Split(accounts, "\n")
	if len(hosts) == 0 || len(lines) != len(hosts) {
		return buserr.New("ErrClusterPasswordMySQLAccounts")
	}
	clauses := make([]string, 0, len(lines))
	for index, line := range lines {
		host, plugin, ok := strings.Cut(line, "\t")
		if !ok || host != hosts[index] || (host != "%" && host != "localhost") {
			return buserr.New("ErrClusterPasswordMySQLAccounts")
		}
		switch plugin {
		case "caching_sha2_password", "mysql_native_password", "sha256_password":
		default:
			return buserr.New("ErrClusterPasswordMySQLAccounts")
		}
		clauses = append(clauses, "'root'@'"+host+"' IDENTIFIED WITH "+plugin+" BY "+quoted)
	}
	query := "SET SESSION sql_mode = 'NO_BACKSLASH_ESCAPES';\nALTER USER " + strings.Join(clauses, ", ") + ";\n"
	if err := db.exec(ctx, io.Discard, strings.NewReader(query), "MYSQL_PWD", password, "mysql", "--host=127.0.0.1",
		"--user=root", "--batch", "--binary-mode", "--default-character-set=utf8mb4"); err != nil {
		return clusterPasswordCommandError(ctx)
	}
	return nil
}

func clusterPasswordCommandError(ctx context.Context) error {
	if ctx.Err() != nil {
		return buserr.New("ErrClusterPasswordUnconfirmed")
	}
	return buserr.New("ErrClusterPasswordApply")
}
