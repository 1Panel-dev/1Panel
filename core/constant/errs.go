package constant

import (
	"errors"
)

const (
	CodeSuccess           = 200
	CodeErrBadRequest     = 400
	CodeErrUnauthorized   = 401
	CodeErrNotFound       = 404
	CodeAuth              = 406
	CodeGlobalLoading     = 407
	CodeErrInternalServer = 500

	CodeErrIP           = 310
	CodeErrDomain       = 311
	CodeErrEntrance     = 312
	CodePasswordExpired = 313

	CodeErrXpack = 410
)

// internal
var (
	ErrCaptchaCode             = errors.New("ErrCaptchaCode")
	ErrAuth                    = errors.New("ErrAuth")
	ErrRecordExist             = errors.New("ErrRecordExist")
	ErrRecordNotFound          = errors.New("ErrRecordNotFound")
	ErrTransform               = errors.New("ErrTransform")
	ErrInitialPassword         = errors.New("ErrInitialPassword")
	ErrInvalidParams           = errors.New("ErrInvalidParams")
	ErrNotSupportType          = errors.New("ErrNotSupportType")
	ErrApiConfigStatusInvalid  = "ErrApiConfigStatusInvalid"
	ErrApiConfigKeyInvalid     = "ErrApiConfigKeyInvalid"
	ErrApiConfigIPInvalid      = "ErrApiConfigIPInvalid"
	ErrApiConfigDisable        = "ErrApiConfigDisable"
	ErrApiConfigKeyTimeInvalid = "ErrApiConfigKeyTimeInvalid"

	ErrTokenParse      = errors.New("ErrTokenParse")
	ErrStructTransform = errors.New("ErrStructTransform")
	ErrPortInUsed      = "ErrPortInUsed"
	ErrCmdTimeout      = "ErrCmdTimeout"
	ErrGroupIsUsed     = "ErrGroupIsUsed"
	ErrGroupIsDefault  = "ErrGroupIsDefault"
)

// api
var (
	ErrTypeInternalServer  = "ErrInternalServer"
	ErrTypeInvalidParams   = "ErrInvalidParams"
	ErrTypeNotLogin        = "ErrNotLogin"
	ErrTypePasswordExpired = "ErrPasswordExpired"
	ErrDemoEnvironment     = "ErrDemoEnvironment"
	ErrEntrance            = "ErrEntrance"
	ErrProxy               = "ErrProxy"
	ErrLocalDelete         = "ErrLocalDelete"

	ErrXpackNotFound    = "ErrXpackNotFound"
	ErrXpackExceptional = "ErrXpackExceptional"
	ErrXpackLost        = "ErrXpackLost"
	ErrXpackTimeout     = "ErrXpackTimeout"
	ErrXpackOutOfDate   = "ErrXpackOutOfDate"
	ErrNoSuchNode       = "ErrNoSuchNode"
	ErrNodeUnbind       = "ErrNodeUnbind"
)

// backup
var (
	ErrBackupInUsed = "ErrBackupInUsed"
	ErrBackupLocal  = "ErrBackupLocal"
	ErrBackupPublic = "ErrBackupPublic"
)

var (
	ErrLicense       = "ErrLicense"
	ErrLicenseCheck  = "ErrLicenseCheck"
	ErrLicenseSave   = "ErrLicenseSave"
	ErrLicenseSync   = "ErrLicenseSync"
	ErrUnbindMaster  = "ErrUnbindMaster"
	ErrFreeNodeLimit = "ErrFreeNodeLimit"
	ErrNodeBound     = "ErrNodeBound"
	ErrNodeBind      = "ErrNodeBind"
	ConnInfoNotMatch = "ConnInfoNotMatch"

	ErrAlertSync = "ErrAlertSync"
)
