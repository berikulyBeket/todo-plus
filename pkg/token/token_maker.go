package token

// TokenMaker defines an interface for managing token creation and validation
type TokenMaker interface {
	CreateToken(userId int) (string, error)
	VerifyToken(tokenString string) (int, error)
}
