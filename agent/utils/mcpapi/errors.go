package mcpapi

import (
	"errors"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/buserr"
)

const (
	BusinessCodeHeader        = "X-Panel-MCP-Business-Code"
	CodeMySQLStopped          = "MCP_MYSQL_STOPPED"
	CodeMySQLConnectionFailed = "MCP_MYSQL_CONNECTION_FAILED"
)

type codedError struct {
	code    string
	message string
}

func (e *codedError) Error() string { return e.message }

func allowedBusinessCode(code string) string {
	switch code {
	case "ErrAliasIsExist", "ErrDomainIsUsed", "ErrPortInUsed", "ErrPortInOtherApp", "ErrPortExist", "ErrRecordExist", "ErrDatabaseIsExist", CodeMySQLStopped, CodeMySQLConnectionFailed:
		return code
	default:
		return ""
	}
}

func BusinessErrorCode(err error) string {
	var coded *codedError
	if errors.As(err, &coded) && coded != nil {
		return allowedBusinessCode(coded.code)
	}
	var business buserr.BusinessError
	if errors.As(err, &business) {
		return allowedBusinessCode(business.Msg)
	}
	var businessPointer *buserr.BusinessError
	if errors.As(err, &businessPointer) && businessPointer != nil {
		return allowedBusinessCode(businessPointer.Msg)
	}
	return ""
}

func NewBusinessError(code, message string) error {
	return &codedError{code: allowedBusinessCode(code), message: message}
}

func SanitizeBusinessError(err error, safeMessage string) error {
	return NewBusinessError(BusinessErrorCode(err), safeMessage)
}

func SetBusinessErrorHeader(request *http.Request, header http.Header, err error) {
	if request == nil || header == nil {
		return
	}
	if present, trusted := Check(request); !present || !trusted {
		return
	}
	header.Del(BusinessCodeHeader)
	if code := BusinessErrorCode(err); code != "" {
		header.Set(BusinessCodeHeader, code)
	}
}
