package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"go-telemetry-server/config"
	"go-telemetry-server/logger"
	"go-telemetry-server/models"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func getPrivateKey() *rsa.PrivateKey {

	privateKeyString := config.Store.JWTInHousePrivateKey

	if privateKeyString == "" {
		logger.Log.Error("JWT Private Key is not configured")
		return &rsa.PrivateKey{}
	}

	block, _ := pem.Decode([]byte(privateKeyString))

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.Log.Error("Can not decode JWT Private Key")
		return &rsa.PrivateKey{}
	}

	// Parse the key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Log.Error("Can not parse PKCS1 JWT Private Key")
	}
	return privateKey

}

func validateJwtInHouse(ctx context.Context, token *jwt.Token, tokenString string) (interface{}, error) {
	logger.Log.Info("Validating JWT token for In-house flow")
	pubKey := getPrivateKey().Public()

	// Check for signing method
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		logger.Log.Error("Invalid JWT token")
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	// Verify token in db
	tokenHash := sha256.Sum256([]byte(tokenString))
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && claims["sub"] != nil {
		userName := claims["sub"].(string)
		if models.IsPersonalAccessTokenExist(
			ctx,
			bson.M{
				"userName":  userName,
				"tokenHash": tokenHash,
			},
		) {
			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}

func validateOIDCToken(_ context.Context, token *jwt.Token, _ string) (interface{}, error) {

	logger.Log.Info("Validating JWT token for OIDC flow")

	// Check for signing method
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		logger.Log.Error("Invalid JWT token")
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok ||
		claims["aud"].(string) != config.Store.Auth.ClientID ||
		claims["iss"].(string) != config.Store.Auth.OidcURL ||
		claims["sub"] == nil {

		return nil, fmt.Errorf("invalid claims")
	}

	return retrievePublicKey(token.Header["kid"].(string))
}

func retrievePublicKey(kid string) (*rsa.PublicKey, error) {
	var retry int
	for retry < 3 {
		pubKey, found := publicKeysMap[kid]
		if found {
			return &pubKey.Key, nil
		}
		logger.Log.Info("Retry retrieving public key from OIDC")
		retry++
		RefreshOIDCInfo()
	}
	return &rsa.PublicKey{}, fmt.Errorf("public key with kid = %v not found", kid)
}

func generateRSAPublicKey(nStr string, eStr string) (rsa.PublicKey, error) {

	decN, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return rsa.PublicKey{}, err
	}
	n := big.NewInt(0)
	n.SetBytes(decN)

	decE, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return rsa.PublicKey{}, err
	}
	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}

	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		return rsa.PublicKey{}, err
	}

	pKey := rsa.PublicKey{
		N: n,
		E: int(e),
	}

	return pKey, nil
}
