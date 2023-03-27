package utils

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/dto"
	"github.com/hoquangnam45/pharmacy-auth/model"
	"github.com/hoquangnam45/pharmacy-common-go/util"
)

func GenerateTokenPair(clientId string, authentication *dto.Authentication) (*util.Pair[string, string], error) {
	return nil, nil
}

func GenerateAccessToken(refreshToken *model.RefreshToken) (string, error) {
	claims := dto.AuthClaims{}
	client := refreshToken.Client
	claims.ExpiresAt = jwt.NewNumericDate(refreshToken.IssuedAt.Add(client.AccessTokenTtl.ToDuration()))
	claims.IssuedAt = jwt.NewNumericDate(refreshToken.IssuedAt)
	claims.ID = uuid.New().String()
	claims.Issuer = client.Issuer
	claims.Subject = refreshToken.Subject
	signingKey := client.SigningKey
	signingMethod := jwt.GetSigningMethod(client.SigningMethod)
	token := jwt.NewWithClaims(signingMethod, claims)
	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(signingKey))
	if err == nil {
		return token.SignedString(privateKey)
	} else {
		return "", err
	}
}

// func (s *Token) RefreshAccessToken(refreshToken string) (string, error) {
// 	refreshToken_, err := s.Verify(refreshToken)
// 	if err != nil {
// 		return "", err
// 	}
// 	now := time.Now()
// 	claims := dto.AuthClaims{}
// 	claims.ID = uuid.New().String()
// 	claims.Issuer = issuer
// 	claims.IssuedAt = &jwt.NumericDate{Time: now}
// 	claims.Subject = subject
// 	claims.ExpiresAt = &jwt.NumericDate{Time: now.Add(ttl)}
// 	claims.NotBefore = &jwt.NumericDate{Time: now}
// 	token := jwt.NewWithClaims(s.signingMethod, claims)
// 	return h.Lift(token.SignedString)(s.signingKey).Eval()
// }

// func (s *Token) Verify(accessToken string) (*jwt.Token, error) {
// algParse := jwt.WithValidMethods([]string{s.signingMethod.Alg()})
// 	token, err := jwt.ParseWithClaims(accessToken, &dto.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(s.verificationKey), nil
// 	}, algParse)
// 	if err != nil {
// 		return nil, err
// 	}
// 	claims := token.Claims.(*dto.AuthClaims)
// 	now := time.Now()
// 	if !claims.VerifyNotBefore(now, false) {
// 		return nil, jwt.ErrTokenUsedBeforeIssued
// 	}
// 	if !claims.VerifyExpiresAt(now, false) {
// 		return nil, jwt.ErrTokenExpired
// 	}
// 	if !claims.VerifyIssuedAt(now, false) {
// 		return nil, jwt.ErrTokenNotValidYet
// 	}
// 	if claims.Issuer != s.issuer {
// 		return nil, jwt.ErrTokenInvalidIssuer
// 	}
// 	if claims.Subject == "" {
// 		return nil, jwt.ErrTokenInvalidId
// 	}
// 	return token, nil
// }
