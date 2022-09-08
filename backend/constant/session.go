package constant

const (
	AuthMethodSession = "session"
	SessionName       = "psession"

	AuthMethodJWT = "jwt"
	JWTHeaderName = "Authorization"
	JWTSigningKey = "1panelKey"
	JWTBufferTime = 86400
	JWTIssuer     = "1Panel"
)
