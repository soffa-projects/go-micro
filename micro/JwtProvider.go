package micro

import (
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/util/dates"
	"time"
)

type TokenProvider interface {
	CreateToken(subject string, issuer string, audience string, claims map[string]string, ttl *time.Duration) (string, error)
	SigningKey() string
}

type DefaultTokenProvider struct {
	secret string
	kind   string
}

func NewJwtTokenProvider(secret string) TokenProvider {
	return &DefaultTokenProvider{secret: secret, kind: "jwt"}
}

func (p *DefaultTokenProvider) SigningKey() string {
	return p.secret
}

func (p *DefaultTokenProvider) CreateToken(subject string, issuer string, audience string, clms map[string]string, ttl *time.Duration) (string, error) {
	if p.kind == "jwt" {
		claims := jwt.MapClaims{}
		claims["sub"] = subject
		if audience != "" {
			claims["aud"] = audience
		}
		if issuer != "" {
			claims["iss"] = issuer
		}
		claims["iat"] = dates.Now().Unix()
		if ttl != nil {
			claims["exp"] = dates.Now().Add(*ttl).Unix()
		}
		if clms != nil {
			for k, v := range clms {
				claims[k] = v
			}
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(p.secret))
		if err != nil {
			log.Errorf("Error signing token: %v", err)
		}
		return signed, err
	} else {
		log.Fatalf("Unsupported token provider: %s", p.kind)
		return "", nil
	}
}
