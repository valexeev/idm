package web

import (
	"fmt"
	jwtMiddleware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"idm/inner/common"
	"os"
	"time"
)

const (
	JwtKey   = "jwt"
	IdmAdmin = "IDM_ADMIN"
	IdmUser  = "IDM_USER"
)

type IdmClaims struct {
	RealmAccess RealmAccessClaims `json:"realm_access"`
	jwt.RegisteredClaims
}

type RealmAccessClaims struct {
	Roles []string `json:"roles"`
}

var CreateAuthMiddleware = func(logger *common.Logger) fiber.Handler {
	if secret := getTestSecret(); secret != "" {
		config := jwtMiddleware.Config{
			ContextKey:   JwtKey,
			ErrorHandler: createJwtErrorHandler(logger),
			Claims:       &IdmClaims{},
			KeyFunc: func(token *jwt.Token) (interface{}, error) {
				if token.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			},
		}
		return jwtMiddleware.New(config)
	}
	config := jwtMiddleware.Config{
		ContextKey:   JwtKey,
		ErrorHandler: createJwtErrorHandler(logger),
		JWKSetURLs:   []string{"http://localhost:9990/realms/idm/protocol/openid-connect/certs"},
		Claims:       &IdmClaims{},
	}
	return jwtMiddleware.New(config)
}

// Для обратной совместимости
var AuthMiddleware = CreateAuthMiddleware

func createJwtErrorHandler(logger *common.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		logger.ErrorCtx(ctx.Context(), "failed autentication", zap.Error(err))
		// Если токен не может быть прочитан, то возвращаем 401
		return common.ErrResponse(
			ctx,
			fiber.StatusUnauthorized,
			err.Error(),
		)
	}
}

// GenerateTestToken генерирует валидный JWT-токен для тестов с заданными ролями
func GenerateTestToken(roles []string) string {
	claims := &IdmClaims{
		RealmAccess: RealmAccessClaims{Roles: roles},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("testsecret"))
	return signed
}

func getTestSecret() string {
	return getEnv("AUTH_TEST_SECRET", "")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// RequireRoles создает middleware для проверки ролей пользователя
func RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем токен из контекста (должен быть установлен AuthMiddleware)
		token, ok := c.Locals(JwtKey).(*jwt.Token)
		if !ok {
			return common.ErrResponse(c, fiber.StatusUnauthorized, "missing or invalid token")
		}

		// Получаем claims из токена
		claims, ok := token.Claims.(*IdmClaims)
		if !ok {
			return common.ErrResponse(c, fiber.StatusUnauthorized, "invalid token claims")
		}

		// Проверяем наличие нужной роли
		hasRole := false
		for _, userRole := range claims.RealmAccess.Roles {
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			return common.ErrResponse(c, fiber.StatusForbidden, "insufficient permissions")
		}

		return c.Next()
	}
}
