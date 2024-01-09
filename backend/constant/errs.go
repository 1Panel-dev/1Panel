package constant

import (
	"errors"
)

const (
	CodeSuccess           = 200
	CodeErrBadRequest     = 400
	CodeErrUnauthorized   = 401
	CodeErrUnSafety       = 402
	CodeErrForbidden      = 403
	CodeErrNotFound       = 404
	CodePasswordExpired   = 405
	CodeAuth              = 406
	CodeGlobalLoading     = 407
	CodeErrIP             = 408
	CodeErrDomain         = 409
	CodeErrInternalServer = 500
	CodeErrHeader         = 406
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
	ErrTypeInternalServer  = "ErrInternalServer"
	ErrTypeInvalidParams   = "ErrInvalidParams"
	ErrTypeNotLogin        = "ErrNotLogin"
	ErrTypePasswordExpired = "ErrPasswordExpired"
	ErrNameIsExist         = "ErrNameIsExist"
	ErrDemoEnvironment     = "ErrDemoEnvironment"
	ErrCmdIllegal          = "ErrCmdIllegal"
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
	ErrFileIsExit       = "ErrFileIsExit"
	ErrFileUpload       = "ErrFileUpload"
	ErrFileDownloadDir  = "ErrFileDownloadDir"
	ErrCmdNotFound      = "ErrCmdNotFound"
	ErrFavoriteExist    = "ErrFavoriteExist"
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
	ErrInUsed       = "ErrInUsed"
	ErrObjectInUsed = "ErrObjectInUsed"
	ErrPortRules    = "ErrPortRules"
	ErrPgImagePull  = "ErrPgImagePull"
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
	ErrFirewall = "ErrFirewall"
)

// cronjob
var (
	ErrBashExecute = "ErrBashExecute"
)

var (
	ErrNotExistUser = "ErrNotExistUser"
)
