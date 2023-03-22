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

	ErrPageGenerate = errors.New("generate page info failed")
	ErrRepoNotValid = "ErrRepoNotValid"
)

// api
var (
	ErrTypeInternalServer  = "ErrInternalServer"
	ErrTypeInvalidParams   = "ErrInvalidParams"
	ErrTypeToken           = "ErrToken"
	ErrTypeTokenTimeOut    = "ErrTokenTimeOut"
	ErrTypeNotLogin        = "ErrNotLogin"
	ErrTypePasswordExpired = "ErrPasswordExpired"
	ErrTypeNotSafety       = "ErrNotSafety"
	ErrNameIsExist         = "ErrNameIsExist"
	ErrDemoEnvironment     = "ErrDemoEnvironment"
)

// app
var (
	ErrPortInUsed          = "ErrPortInUsed"
	ErrAppLimit            = "ErrAppLimit"
	ErrAppRequired         = "ErrAppRequired"
	ErrFileCanNotRead      = "ErrFileCanNotRead"
	ErrFileToLarge         = "ErrFileToLarge"
	ErrNotInstall          = "ErrNotInstall"
	ErrPortInOtherApp      = "ErrPortInOtherApp"
	ErrDbUserNotValid      = "ErrDbUserNotValid"
	ErrUpdateBuWebsite     = "ErrUpdateBuWebsite"
	Err1PanelNetworkFailed = "Err1PanelNetworkFailed"
)

//website
var (
	ErrDomainIsExist = "ErrDomainIsExist"
	ErrAliasIsExist  = "ErrAliasIsExist"
	ErrAppDelete     = "ErrAppDelete"
	ErrGroupIsUsed   = "ErrGroupIsUsed"
)

//ssl
var (
	ErrSSLCannotDelete     = "ErrSSLCannotDelete"
	ErrAccountCannotDelete = "ErrAccountCannotDelete"
	ErrSSLApply            = "ErrSSLApply"
	ErrEmailIsExist        = "ErrEmailIsExist"
)

//file
var (
	ErrPathNotFound     = "ErrPathNotFound"
	ErrMovePathFailed   = "ErrMovePathFailed"
	ErrLinkPathNotFound = "ErrLinkPathNotFound"
	ErrFileIsExit       = "ErrFileIsExit"
	ErrFileUpload       = "ErrFileUpload"
)

//mysql
var (
	ErrUserIsExist     = "ErrUserIsExist"
	ErrDatabaseIsExist = "ErrDatabaseIsExist"
)

//redis
var (
	ErrTypeOfRedis = "ErrTypeOfRedis"
)

//container
var (
	ErrInUsed       = "ErrInUsed"
	ErrObjectInUsed = "ErrObjectInUsed"
)
