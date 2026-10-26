package mcpapi

import (
	"errors"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/utils/mcpapi/metadata"
)

const (
	BusinessCodeHeader        = metadata.BusinessCodeHeader
	CodeMySQLStopped          = "MCP_MYSQL_STOPPED"
	CodeMySQLConnectionFailed = "MCP_MYSQL_CONNECTION_FAILED"

	WebsiteAliasHint   = "1Panel reports that the website identifier already exists."
	WebsiteDomainHint  = "1Panel reports that the domain is already used by a website."
	PortOccupiedHint   = "1Panel reports that the requested port is already in use."
	DatabaseExistsHint = "1Panel reports that the database or its panel record already exists."
)

// responseBusinessCode accepts one known code from the internal Core/Agent
// response. Localized messages and arbitrary metadata are never used as hints.
func responseBusinessCode(header http.Header) string {
	values := header.Values(BusinessCodeHeader)
	if len(values) != 1 || len(values[0]) > 64 {
		return ""
	}
	switch values[0] {
	case "ErrAliasIsExist", "ErrDomainIsUsed", "ErrPortInUsed", "ErrPortInOtherApp", "ErrPortExist", "ErrRecordExist", "ErrDatabaseIsExist", "ErrProxy", CodeMySQLStopped, CodeMySQLConnectionFailed:
		return values[0]
	}
	return ""
}

func BusinessErrorHint(method, path, businessCode string) string {
	if method != http.MethodPost {
		return ""
	}
	switch path {
	case "/api/v2/websites":
		switch businessCode {
		case "ErrAliasIsExist":
			return WebsiteAliasHint
		case "ErrDomainIsUsed":
			return WebsiteDomainHint
		}
	case "/api/v2/apps/install":
		switch businessCode {
		case "ErrPortInUsed", "ErrPortInOtherApp", "ErrPortExist":
			return PortOccupiedHint
		}
	case "/api/v2/databases":
		switch businessCode {
		case "ErrRecordExist", "ErrDatabaseIsExist":
			return DatabaseExistsHint
		}
	}
	return ""
}

type businessHintError struct{ hint string }

func (e *businessHintError) Error() string { return e.hint }

func SafeBusinessErrorMessage(err error) string {
	var failure *businessHintError
	if !errors.As(err, &failure) || failure == nil {
		return ""
	}
	switch failure.hint {
	case WebsiteAliasHint, WebsiteDomainHint, PortOccupiedHint, DatabaseExistsHint:
		return failure.hint + " Operation outcome is unknown; check the target resource in 1Panel before retrying"
	}
	return ""
}
