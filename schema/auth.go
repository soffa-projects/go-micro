package schema

type AuthToken struct {
	Issuer   string `json:"token"`
	Audience string `json:"audience"`
}

type Authentication struct {
	Token         *AuthToken
	Authenticated bool
	Name          string
	Email         string
	UserId        string
	Roles         []string
	Permissions   []string
}