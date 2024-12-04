package handler

import (
	"context"
	"fmt"
	"go-graphql-user-svc/config"
	"go-graphql-user-svc/internal/repository"
	"go-graphql-user-svc/internal/service"
	"go-graphql-user-svc/util"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Allow any origin (you can restrict this to specific domains)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Allow specific methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Allow specific headers

		// Handle preflight requests (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func getJWTMiddleWare(cfg *config.Config) func(http.Handler) http.Handler {
	var jwtKey = []byte(cfg.AuthTokenConfig.SecretKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate the token signing method (use RS256, HS256, etc.)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Token is valid, add it to the context
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Optionally, you can add user data from claims into the context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "claims", claims)

			// Proceed with the next handler, passing the updated context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func SetupServer() {
	cfg := config.GetConfig()
	db := util.GetMongoDB(cfg)
	claimsValidator := service.NewClaimsValidator()
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)
	userHandler := NewUserHandler(userService, claimsValidator)
	authHandler := NewAuthHandler(userService)
	jwtMiddleware := getJWTMiddleWare(cfg)
	http.Handle("/user-svc", corsMiddleware(jwtMiddleware(http.HandlerFunc(userHandler.ServeGraphQL))))
	http.Handle("/user-svc/auth", corsMiddleware(http.HandlerFunc(authHandler.Handle)))

	var serverMsg = fmt.Sprintf(
		"GraphQL server running at http://localhost:%v/user-svc",
		cfg.ServerConfig.Port,
	)
	log.Println(serverMsg)

	addressFmt := "%v:%v"
	address := fmt.Sprintf(addressFmt, cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	log.Fatal(http.ListenAndServe(address, nil))
}
