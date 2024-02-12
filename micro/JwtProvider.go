package micro

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/util/dates"
	"github.com/thoas/go-funk"
	"time"
)

type TokenProvider interface {
	CreateToken(subject string, issuer string, audience string, claims map[string]interface{}, ttl time.Duration) (string, error)
	Decode(token string, checkSignature bool) (map[string]interface{}, error)
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

func (p *DefaultTokenProvider) Decode(token string, checkSignature bool) (map[string]interface{}, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(p.secret), nil
	})

	if err != nil && checkSignature {
		return nil, errors.New("invalid_signature")
	}
	if checkSignature && !t.Valid {
		return nil, errors.New("invalid_token")
	}
	res := make(map[string]interface{})
	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		// Convert MapClaims to map[string]interface{}
		for key, value := range claims {
			res[key] = value
		}
	}
	return res, nil
}

func (p *DefaultTokenProvider) CreateToken(subject string, issuer string, audience string, clms map[string]interface{}, ttl time.Duration) (string, error) {
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
		if ttl != time.Duration(0) {
			claims["exp"] = dates.Now().Add(ttl).Unix()
		}
		if clms != nil {
			for k, v := range clms {
				if !funk.IsEmpty(v) {
					claims[k] = v
				}
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
