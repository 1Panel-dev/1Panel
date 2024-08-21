package constant

import (
	"errors"
)

const (
	CodeSuccess           = 200
	CodeErrBadRequest     = 400
	CodeGlobalLoading     = 407
	CodeErrInternalServer = 500

	CodeErrXpack = 410
)

// internal
var (
	ErrRecordExist     = errors.New("ErrRecordExist")
	ErrRecordNotFound  = errors.New("ErrRecordNotFound")
	ErrStructTransform = errors.New("ErrStructTransform")
	ErrNotSupportType  = errors.New("ErrNotSupportType")
	ErrInvalidParams   = errors.New("ErrInvalidParams")
)

// api
var (
	ErrTypeInternalServer = "ErrInternalServer"
	ErrTypeInvalidParams  = "ErrInvalidParams"
	ErrNameIsExist        = "ErrNameIsExist"
	ErrDemoEnvironment    = "ErrDemoEnvironment"
	ErrCmdIllegal         = "ErrCmdIllegal"
	ErrXpackNotFound      = "ErrXpackNotFound"
	ErrXpackNotActive     = "ErrXpackNotActive"
	ErrXpackOutOfDate     = "ErrXpackOutOfDate"
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

var (
	ErrNotExistUser = "ErrNotExistUser"
)

// license
var (
	ErrLicense      = "ErrLicense"
	ErrLicenseCheck = "ErrLicenseCheck"
	ErrLicenseSave  = "ErrLicenseSave"
	ErrLicenseSync  = "ErrLicenseSync"
)
