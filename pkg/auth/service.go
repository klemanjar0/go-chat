package auth

// Service bundles the password and refresh-token primitives behind a single
// value type so callers can depend on a CryptoService-style interface instead
// of importing this package directly. The package-level functions remain
// available for code that genuinely wants the primitives without an indirection.
type Service struct{}

func (Service) HashPassword(password string, cost int) (string, error) {
	return HashPassword(password, cost)
}

func (Service) VerifyPassword(hash, password string) error {
	return VerifyPassword(hash, password)
}

func (Service) GenerateRefreshToken() (string, error) {
	return GenerateRefreshToken()
}

func (Service) HashRefreshToken(token string) string {
	return HashRefreshToken(token)
}
