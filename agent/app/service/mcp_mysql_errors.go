package service

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
)

const mcpMysqlStoppedMessage = "The selected MySQL instance is not running. Start the database service and try the query again."
const mcpMysqlConnectionFailedMessage = "Cannot connect to the selected MySQL instance. Check the database service, connection settings and panel logs."

var errMCPMysqlStopped = mcpapi.NewBusinessError(mcpapi.CodeMySQLStopped, mcpMysqlStoppedMessage)
var errMCPMysqlConnectionFailed = mcpapi.NewBusinessError(mcpapi.CodeMySQLConnectionFailed, mcpMysqlConnectionFailedMessage)

var mcpMysqlStoppedPattern = regexp.MustCompile(`^Error response from daemon: [Cc]ontainer [A-Za-z0-9][A-Za-z0-9_.-]{0,127} is not running$`)
var mcpMysqlConnectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^ERROR 2002 \(HY000\): Can't connect to local MySQL server through socket '[^'\r\n]{1,1000}' \([0-9]{1,10}\)$`),
	regexp.MustCompile(`^ERROR 2003 \(HY000\): Can't connect to MySQL server on '[^'\r\n]{1,1000}' \([0-9]{1,10}\)$`),
	regexp.MustCompile(`^ERROR 1045 \(28000\): Access denied for user '[^'\r\n]{1,128}'@'[^'\r\n]{1,255}' \(using password: (YES|NO)\)$`),
}

func classifyMCPMysqlReadFailure(stderr string) error {
	generic := errors.New("local MySQL status query failed")
	if len(stderr) == 0 || len(stderr) > 2048 || !utf8.ValidString(stderr) {
		return generic
	}
	warning := "mysql: [Warning] Using a password on the command line interface can be insecure."
	stderr = strings.TrimPrefix(stderr, warning+"\n")
	if strings.HasSuffix(stderr, "\r\n") {
		stderr = strings.TrimSuffix(stderr, "\r\n")
	} else {
		stderr = strings.TrimSuffix(stderr, "\n")
	}
	if strings.IndexFunc(stderr, unicode.IsControl) >= 0 {
		return generic
	}
	if mcpMysqlStoppedPattern.MatchString(stderr) {
		return errMCPMysqlStopped
	}
	for _, pattern := range mcpMysqlConnectionPatterns {
		if pattern.MatchString(stderr) {
			return errMCPMysqlConnectionFailed
		}
	}
	return generic
}
