package ports

//go:generate mockgen -source token.go -destination mocks/mock_token.go -package mocks
type TokenProvider interface {
	Create(userID string) (string, error)
	Parse(tokenString string) (string, error)
}
