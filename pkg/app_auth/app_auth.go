package appauth

// Interface defines the contract for app authentication logic
type Interface interface {
	Validate(appId, appKey, accessType string) bool
}

// Constants representing access types
const (
	PrivateAccess = "private"
	PublicAccess  = "public"
)

// AppAuth holds credentials for both public and private access
type AppAuth struct {
	appId         string
	appKey        string
	privateAppId  string
	privateAppKey string
}

// New creates a new instance of AppAuth with public and private credentials
func New(appId, appKey, privateAppId, privateAppKey string) Interface {
	return &AppAuth{
		appId:         appId,
		appKey:        appKey,
		privateAppId:  privateAppId,
		privateAppKey: privateAppKey,
	}
}

// Validate checks if the provided appId and appKey match either the public or private credentials based on the access type
func (a *AppAuth) Validate(appId, appKey, accessType string) bool {
	switch accessType {
	case PrivateAccess:
		return a.privateAppId == appId && a.privateAppKey == appKey
	case PublicAccess:
		return a.appId == appId && a.appKey == appKey
	default:
		return false
	}
}
