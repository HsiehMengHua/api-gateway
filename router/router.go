package router

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Setup() *gin.Engine {
	r := gin.Default()
	r.StaticFile("/version", "./version.txt")

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
		}

		userApi.POST("", func(c *gin.Context) {
			proxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Payment API Proxy
	{
		paymentApi := r.Group("/payments")
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
			handler(c, proxy)
		})

		paymentApi.POST("/withdraw", func(c *gin.Context) {
			handler(c, proxy)
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
