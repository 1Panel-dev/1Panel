package mcpapi

import "errors"

const (
	NodeAbnormalMessage          = "1Panel reports an abnormal status for the selected node. Check the node status, Agent service and panel logs."
	MySQLStoppedMessage          = "The selected MySQL instance is not running. Start the database service and try the query again."
	MySQLConnectionFailedMessage = "Cannot connect to the selected MySQL instance. Check the database service, connection settings and panel logs."
	MySQLStatusFailedMessage     = "Cannot read the selected MySQL instance status. Check that the database service is running and review its connection settings and panel logs."
	SystemInfoFailedMessage      = "Cannot read system information from the selected 1Panel node. Check the Agent service and panel logs."
)

func ReadErrorHint(method, fullAPIPath, businessCode string) string {
	if method == "POST" && fullAPIPath == "/api/v2/databases/status" {
		switch businessCode {
		case CodeMySQLStopped:
			return MySQLStoppedMessage
		case CodeMySQLConnectionFailed:
			return MySQLConnectionFailedMessage
		}
	}
	return ""
}

type readResponseError struct {
	message         string
	systemTransport bool
}

func (e *readResponseError) Error() string { return e.message }

func SafeReadErrorMessage(err error) string {
	var failure *readResponseError
	if !errors.As(err, &failure) || failure == nil {
		return ""
	}
	switch failure.message {
	case MySQLStoppedMessage, MySQLConnectionFailedMessage, MySQLStatusFailedMessage, SystemInfoFailedMessage, NodeAbnormalMessage:
		return failure.message
	}
	return ""
}

func IsSystemReadTransportFailure(err error) bool {
	var failure *readResponseError
	return errors.As(err, &failure) && failure != nil && failure.message == SystemInfoFailedMessage && failure.systemTransport
}

func classifyReadResponseFailure(input Request, status, code int, businessCode string, parsed bool) error {
	switch {
	case input.Method == "POST" && input.Path == "/api/v2/databases/status":
		if !parsed || (code != 400 && code != 500) || (status != 200 && status != 400 && (status < 500 || status > 599)) {
			return nil
		}
		hint := ReadErrorHint(input.Method, input.Path, businessCode)
		if hint == "" {
			hint = MySQLStatusFailedMessage
		}
		return &readResponseError{message: hint}
	case input.Method == "GET" && input.Path == "/api/v2/dashboard/base/os":
		if (status >= 400 && status < 500) || (parsed && code >= 400 && code < 500) {
			return nil
		}
		serverStatus := status >= 500 && status <= 599
		businessFailure := parsed && status >= 200 && status < 300 && code >= 500 && code <= 599
		if !serverStatus && !businessFailure {
			return nil
		}
		if parsed && code != 0 && code != 200 && (code < 500 || code > 599) {
			return nil
		}
		transport := status == 502 || status == 503 || status == 504 || (parsed && code == 500 && businessCode == "ErrProxy")
		return &readResponseError{message: SystemInfoFailedMessage, systemTransport: transport}
	}
	return nil
}
