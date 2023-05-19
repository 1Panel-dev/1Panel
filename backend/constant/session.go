package constant

const (
	AuthMethodSession = "session"
	SessionName       = "psession"

	AuthMethodJWT = "jwt"
	JWTHeaderName = "PanelAuthorization"
	JWTBufferTime = 3600
	JWTIssuer     = "1Panel"

	PasswordExpiredName = "expired"
)
