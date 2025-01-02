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
	ErrCaptchaCode     = errors.New("ErrCaptchaCode")
	ErrAuth            = errors.New("ErrAuth")
	ErrRecordExist     = errors.New("ErrRecordExist")
	ErrRecordNotFound  = errors.New("ErrRecordNotFound")
	ErrStructTransform = errors.New("ErrStructTransform")
	ErrInitialPassword = errors.New("ErrInitialPassword")
	ErrNotSupportType  = errors.New("ErrNotSupportType")
	ErrInvalidParams   = errors.New("ErrInvalidParams")

	ErrTokenParse = errors.New("ErrTokenParse")
)

// api
var (
	ErrTypeInternalServer      = "ErrInternalServer"
	ErrTypeInvalidParams       = "ErrInvalidParams"
	ErrTypeNotLogin            = "ErrNotLogin"
	ErrTypePasswordExpired     = "ErrPasswordExpired"
	ErrNameIsExist             = "ErrNameIsExist"
	ErrDemoEnvironment         = "ErrDemoEnvironment"
	ErrCmdIllegal              = "ErrCmdIllegal"
	ErrXpackNotFound           = "ErrXpackNotFound"
	ErrXpackNotActive          = "ErrXpackNotActive"
	ErrXpackLost               = "ErrXpackLost"
	ErrXpackTimeout            = "ErrXpackTimeout"
	ErrXpackOutOfDate          = "ErrXpackOutOfDate"
	ErrApiConfigStatusInvalid  = "ErrApiConfigStatusInvalid"
	ErrApiConfigKeyInvalid     = "ErrApiConfigKeyInvalid"
	ErrApiConfigIPInvalid      = "ErrApiConfigIPInvalid"
	ErrApiConfigDisable        = "ErrApiConfigDisable"
	ErrApiConfigKeyTimeInvalid = "ErrApiConfigKeyTimeInvalid"
)

// app
var (
	ErrPortInUsed          = "ErrPortInUsed"
	ErrAppLimit            = "ErrAppLimit"
	ErrFileCanNotRead      = "ErrFileCanNotRead"
	ErrNotInstall          = "ErrNotInstall"
	ErrPortInOtherApp      = "ErrPortInOtherApp"
	ErrDbUserNotValid      = "ErrDbUserNotValid"
	ErrUpdateBuWebsite     = "ErrUpdateBuWebsite"
	Err1PanelNetworkFailed = "Err1PanelNetworkFailed"
	ErrCmdTimeout          = "ErrCmdTimeout"
	ErrFileParse           = "ErrFileParse"
	ErrInstallDirNotFound  = "ErrInstallDirNotFound"
	ErrContainerName       = "ErrContainerName"
	ErrAppNameExist        = "ErrAppNameExist"
	ErrFileNotFound        = "ErrFileNotFound"
	ErrFileParseApp        = "ErrFileParseApp"
	ErrAppParamKey         = "ErrAppParamKey"
)

// website
var (
	ErrDomainIsExist      = "ErrDomainIsExist"
	ErrAliasIsExist       = "ErrAliasIsExist"
	ErrGroupIsUsed        = "ErrGroupIsUsed"
	ErrUsernameIsExist    = "ErrUsernameIsExist"
	ErrUsernameIsNotExist = "ErrUsernameIsNotExist"
	ErrBackupMatch        = "ErrBackupMatch"
	ErrBackupExist        = "ErrBackupExist"
	ErrDomainIsUsed       = "ErrDomainIsUsed"
)

// ssl
var (
	ErrSSLCannotDelete               = "ErrSSLCannotDelete"
	ErrAccountCannotDelete           = "ErrAccountCannotDelete"
	ErrSSLApply                      = "ErrSSLApply"
	ErrEmailIsExist                  = "ErrEmailIsExist"
	ErrEabKidOrEabHmacKeyCannotBlank = "ErrEabKidOrEabHmacKeyCannotBlank"
)

// file
var (
	ErrPathNotFound     = "ErrPathNotFound"
	ErrMovePathFailed   = "ErrMovePathFailed"
	ErrLinkPathNotFound = "ErrLinkPathNotFound"
	ErrFileIsExist      = "ErrFileIsExist"
	ErrFileUpload       = "ErrFileUpload"
	ErrFileDownloadDir  = "ErrFileDownloadDir"
	ErrCmdNotFound      = "ErrCmdNotFound"
	ErrFavoriteExist    = "ErrFavoriteExist"
	ErrPathNotDelete    = "ErrPathNotDelete"
)

// mysql
var (
	ErrUserIsExist     = "ErrUserIsExist"
	ErrDatabaseIsExist = "ErrDatabaseIsExist"
	ErrExecTimeOut     = "ErrExecTimeOut"
	ErrRemoteExist     = "ErrRemoteExist"
	ErrLocalExist      = "ErrLocalExist"
)

// redis
var (
	ErrTypeOfRedis = "ErrTypeOfRedis"
)

// container
var (
	ErrInUsed            = "ErrInUsed"
	ErrObjectInUsed      = "ErrObjectInUsed"
	ErrObjectBeDependent = "ErrObjectBeDependent"
	ErrPortRules         = "ErrPortRules"
	ErrPgImagePull       = "ErrPgImagePull"
)

// runtime
var (
	ErrDirNotFound         = "ErrDirNotFound"
	ErrFileNotExist        = "ErrFileNotExist"
	ErrImageBuildErr       = "ErrImageBuildErr"
	ErrImageExist          = "ErrImageExist"
	ErrDelWithWebsite      = "ErrDelWithWebsite"
	ErrRuntimeStart        = "ErrRuntimeStart"
	ErrPackageJsonNotFound = "ErrPackageJsonNotFound"
	ErrScriptsNotFound     = "ErrScriptsNotFound"
)

var (
	ErrBackupInUsed = "ErrBackupInUsed"
	ErrOSSConn      = "ErrOSSConn"
	ErrEntrance     = "ErrEntrance"
)

var (
	ErrFirewallNone = "ErrFirewallNone"
	ErrFirewallBoth = "ErrFirewallBoth"
)

// cronjob
var (
	ErrBashExecute = "ErrBashExecute"
)

var (
	ErrNotExistUser = "ErrNotExistUser"
)

// license
var (
	ErrLicense      = "ErrLicense"
	ErrLicenseCheck = "ErrLicenseCheck"
	ErrXpackVersion = "ErrXpackVersion"
	ErrLicenseSave  = "ErrLicenseSave"
	ErrLicenseSync  = "ErrLicenseSync"
)

// alert
var (
	ErrAlert       = "ErrAlert"
	ErrAlertPush   = "ErrAlertPush"
	ErrAlertSave   = "ErrAlertSave"
	ErrAlertSync   = "ErrAlertSync"
	ErrAlertRemote = "ErrAlertRemote"
)
