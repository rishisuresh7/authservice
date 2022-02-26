package helper

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	mr "math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/sirupsen/logrus"

	"authservice/models"
	"authservice/repository"
)

type Helper interface {
	UnMarshal(data []byte, dest interface{})
	Marshal(src interface{}) []byte
	SendOTP(ctx context.Context, phone string) (string, error)
	GetJWT(userClaims map[string]interface{}) (string, error)
	DecodeJWT(token string) (map[string]interface{}, error)
	EncodeClaims(userClaims map[string]interface{}) (string, error)
	DecodeToken(data string) (*models.RefreshMeta, error)
	NewId() string
}

type helper struct {
	tokenSecret   string
	refreshSecret string

	logger *logrus.Logger
	redis  repository.RedisQueryer
}

func NewHelper(l *logrus.Logger, r repository.RedisQueryer, ts, rs string) Helper {
	return &helper{
		redis:         r,
		logger:        l,
		tokenSecret:   ts,
		refreshSecret: rs,
	}
}

func (h *helper) UnMarshal(data []byte, dest interface{}) {
	_ = json.Unmarshal(data, dest)
}

func (h *helper) Marshal(src interface{}) []byte {
	res, _ := json.Marshal(src)
	return res
}

func (h *helper) SendOTP(ctx context.Context, phone string) (string, error) {
	// key, otp := h.generateOTPKeyPair(6)
	// message := fmt.Sprintf("%s is your NEERBI authentication code.", otp)
	// go func() {
	// 	_, err := h.aws.Publish(&sns.PublishInput{
	// 		Message:     &message,
	// 		PhoneNumber: &phone,
	// 	})
	// 	if err != nil {
	// 		h.logger.Errorf("SendOTP: unable to publish OTP: %s", err)
	// 	}
	// }()

	// err := h.redis.Set(ctx, key, otp, 60*time.Second)
	// if err != nil {
	// 	return "", fmt.Errorf("sendOTP: unable to save OTP: %s", err)
	// }

	// return key, nil
	return "", nil
}

func (h *helper) generateOTPKeyPair(digits int) (string, string) {
	numbers := [10]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	b := make([]byte, digits)
	n, err := io.ReadAtLeast(rand.Reader, b, digits)
	if n != digits || err != nil {
		for i := 0; i < digits; i++ {
			b[i] = numbers[mr.Intn(10)]
		}
	} else {
		for i := 0; i < len(b); i++ {
			b[i] = numbers[int(b[i])%10]
		}
	}

	key, _ := uuid.NewV4()
	return key.String(), string(b)
}

func (h *helper) GetJWT(userClaims map[string]interface{}) (string, error) {
	claims := jwt.MapClaims(userClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 20).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString([]byte(h.tokenSecret))
	if err != nil {
		return "", fmt.Errorf("getJWT: unable to sign token: %s", err)
	}

	return fmt.Sprintf("Bearer %s", signedToken), nil
}

func (h *helper) DecodeJWT(token string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	decodedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(h.tokenSecret), nil
	})
	if err != nil {
		return claims, fmt.Errorf("decodeJWT: unable to decode JWT: %s", err)
	}

	if !decodedToken.Valid {
		return claims, fmt.Errorf("decodeJWT: invalid JWT")
	}

	return claims, nil
}

func (h *helper) EncodeClaims(userClaims map[string]interface{}) (string, error) {
	dataBytes := h.Marshal(&models.RefreshMeta{
		UserClaims: userClaims,
		Expiry:     time.Now().Add(time.Hour * 24 * 5).Unix(),
	})
	refreshToken, err := h.encrypt(h.refreshSecret, dataBytes)
	if err != nil {
		return "", fmt.Errorf("encodeClaims: unable to encrypt data: %s", err)
	}

	return hex.EncodeToString(refreshToken), nil
}

func (h *helper) DecodeToken(data string) (*models.RefreshMeta, error) {
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decodeToken: cannot decode from hex: %s", err)
	}

	claimBytes, err := h.decrypt(h.refreshSecret, dataBytes)
	if err != nil {
		return nil, fmt.Errorf("decodeToken: unable to decode token: %s", err)
	}

	var claims models.RefreshMeta
	h.UnMarshal(claimBytes, &claims)

	return &claims, nil
}

func (h *helper) hash(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func (h *helper) encrypt(key string, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(h.hash(key))
	if err != nil {
		return nil, fmt.Errorf("encrypt: unable to create block: %s", err)
	}

	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("encrypt: unable to read buffer: %s", err)
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

	return ciphertext, nil
}

func (h *helper) decrypt(key string, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(h.hash(key))
	if err != nil {
		return nil, fmt.Errorf("decrypt: unable to create block: %s", err)
	}

	if len(text) < aes.BlockSize {
		return nil, fmt.Errorf("decrypt: cypher text too short")
	}

	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, fmt.Errorf("decrypt: unable to decode data: %s", err)
	}

	return data, nil
}

func (h *helper) NewId() string {
	uid, _ := uuid.NewV4()
	return uid.String()
}
