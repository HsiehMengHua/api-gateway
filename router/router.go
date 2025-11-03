package router

import (
	"api-gateway/middlewares"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.StaticFile("/version", "./version.txt")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://localhost:5173", "https://payment-frontend-production.up.railway.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// User API Proxy
	{
		userApi := r.Group("/user")
		target, _ := url.Parse(os.Getenv("USER_SERVICE_HOST"))
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.Director = func(req *http.Request) {
			req.Host = target.Host
			req.URL.Host = target.Host
			req.URL.Scheme = target.Scheme

			if req.URL.Path == "/user" {
				req.URL.Path = "/api/v1/user"
			}

			if req.URL.Path == "/user/login" {
				req.URL.Path = "/api/v1/user/login"
			}
		}

		userApi.POST("", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})

		userApi.POST("/login", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Payment API Proxy
	{
		paymentApi := r.Group("/payments")
		paymentApi.Use(middlewares.RequireAuthorization())

		target, _ := url.Parse(os.Getenv("PAYMENT_SERVICE_HOST"))
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.Director = func(req *http.Request) {
			req.Host = target.Host
			req.URL.Host = target.Host
			req.URL.Scheme = target.Scheme

			if req.URL.Path == "/payments/deposit" {
				req.URL.Path = "/api/v1/payments/deposit"
			}

			if req.URL.Path == "/payments/withdraw" {
				req.URL.Path = "/api/v1/payments/withdraw"
			}

			if req.URL.Path == "/payments/transfer" {
				req.URL.Path = "/api/v1/payments/transfer"
			}

			if req.URL.Path == "/payments/confirm" {
				req.URL.Path = "/api/v1/payments/confirm"
			}

			if req.URL.Path == "/payments/cancel" {
				req.URL.Path = "/api/v1/payments/cancel"
			}
		}

		paymentApi.POST("/deposit", func(c *gin.Context) {
			handlerWithUserId(c, proxy)
		})

		paymentApi.POST("/withdraw", func(c *gin.Context) {
			handlerWithUserId(c, proxy)
		})

		paymentApi.POST("/transfer", func(c *gin.Context) {
			handler(c, proxy)
		})

		paymentApi.POST("/confirm", func(c *gin.Context) {
			handler(c, proxy)
		})

		paymentApi.POST("/cancel", func(c *gin.Context) {
			handler(c, proxy)
		})
	}

	return r
}

func handler(c *gin.Context, proxy *httputil.ReverseProxy) {
	var scheme string
	if c.Request.TLS != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	log.Infof("Forwarding to `%s%s%s`", scheme, c.Request.Host, c.Request.RequestURI)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func handlerWithUserId(c *gin.Context, proxy *httputil.ReverseProxy) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}

	var userIdStr string
	switch v := userId.(type) {
	case string:
		userIdStr = v
	case float64:
		userIdStr = fmt.Sprintf("%.0f", v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return
	}

	c.Request.Header.Set("X-User-Id", userIdStr)

	var scheme string
	if c.Request.TLS != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	log.Infof("Forwarding to `%s%s%s` with user_id: %v", scheme, c.Request.Host, c.Request.RequestURI, userId)

	proxy.ServeHTTP(c.Writer, c.Request)
}
