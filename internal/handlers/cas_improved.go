package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"
)

// CASTicketManager manages CAS service tickets using go-cache
type CASTicketManager struct {
	cache *cache.Cache
}

// CASTicket represents a CAS service ticket
type CASTicket struct {
	Username  string    `json:"username"`
	Service   string    `json:"service"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

var (
	casTicketManager *CASTicketManager
)

// InitCASImproved initializes improved CAS implementation with go-cache
func InitCASImproved() error {
	// Create ticket manager with 5 minute default expiration and 10 minute cleanup interval
	casTicketManager = &CASTicketManager{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	logger.Info("CAS initialized successfully with improved implementation using go-cache")
	return nil
}

// GenerateServiceTicket generates a new CAS service ticket
func (tm *CASTicketManager) GenerateServiceTicket(username, service string) (string, error) {
	// Generate ticket ID
	ticketBytes := make([]byte, 32)
	if _, err := rand.Read(ticketBytes); err != nil {
		return "", err
	}
	ticketID := "ST-" + base64.URLEncoding.EncodeToString(ticketBytes)

	// Create ticket
	ticket := &CASTicket{
		Username:  username,
		Service:   service,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	// Store in cache with 5 minute expiration
	tm.cache.Set(ticketID, ticket, 5*time.Minute)

	// Also store in database for persistence
	dbTicket := CASServiceTicket{
		Ticket:    ticketID,
		Service:   service,
		Username:  username,
		ExpiresAt: ticket.ExpiresAt,
		Used:      false,
	}
	dbTicket.ID = utils.GenerateTradeIDString("tkt")

	if err := database.DB.Create(&dbTicket).Error; err != nil {
		logger.Error("Failed to save CAS ticket to database", zap.Error(err))
		// Continue anyway, as we have it in cache
	}

	logger.Info("Generated CAS service ticket",
		zap.String("ticket", ticketID),
		zap.String("service", service),
		zap.String("username", username),
	)

	return ticketID, nil
}

// ValidateServiceTicket validates a CAS service ticket
func (tm *CASTicketManager) ValidateServiceTicket(ticketID, service string) (string, bool) {
	// First try cache
	if item, found := tm.cache.Get(ticketID); found {
		ticket := item.(*CASTicket)

		// Check if ticket matches service and hasn't been used
		if ticket.Service == service && !ticket.Used && time.Now().Before(ticket.ExpiresAt) {
			// Mark as used
			ticket.Used = true
			tm.cache.Set(ticketID, ticket, cache.DefaultExpiration)

			// Also update database
			database.DB.Model(&CASServiceTicket{}).Where("ticket = ?", ticketID).Update("used", true)

			return ticket.Username, true
		}
	}

	// Fallback to database
	var dbTicket CASServiceTicket
	if err := database.DB.Where("ticket = ? AND service = ? AND used = ? AND expires_at > ?",
		ticketID, service, false, time.Now()).First(&dbTicket).Error; err != nil {
		return "", false
	}

	// Mark as used
	dbTicket.Used = true
	database.DB.Save(&dbTicket)

	// Update cache
	cacheTicket := &CASTicket{
		Username:  dbTicket.Username,
		Service:   dbTicket.Service,
		CreatedAt: dbTicket.CreatedAt,
		ExpiresAt: dbTicket.ExpiresAt,
		Used:      true,
	}
	tm.cache.Set(ticketID, cacheTicket, cache.DefaultExpiration)

	return dbTicket.Username, true
}

// CASLoginHandlerImproved handles CAS login with improved implementation
func CASLoginHandlerImproved(c *gin.Context) {
	service := c.Query("service")
	gateway := c.Query("gateway") == "true"
	renew := c.Query("renew") == "true"

	logger.Info("CAS login request (improved)",
		zap.String("service", service),
		zap.Bool("gateway", gateway),
		zap.Bool("renew", renew),
		zap.String("ip", c.ClientIP()),
	)

	// Check if user is already authenticated
	user, authenticated := checkUserAuthenticationFromSession(c)
	if authenticated && !renew {
		// User is already authenticated, generate service ticket
		if service != "" {
			ticket, err := casTicketManager.GenerateServiceTicket(user.Username, service)
			if err != nil {
				logger.Error("Failed to generate CAS service ticket", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate service ticket",
				})
				return
			}

			// Redirect back to service with ticket
			redirectURL := buildCASRedirectURL(service, ticket)
			c.Redirect(http.StatusFound, redirectURL)
			return
		}

		// No service specified, show success page
		c.JSON(http.StatusOK, gin.H{
			"message": "Already authenticated",
			"user":    user.Username,
		})
		return
	}

	if gateway {
		// Gateway mode - redirect back to service without authentication
		if service != "" {
			c.Redirect(http.StatusFound, service)
			return
		}
	}

	// Show login form
	c.HTML(http.StatusOK, "cas_login.html", gin.H{
		"service": service,
		"gateway": gateway,
		"renew":   renew,
		"title":   "CAS Login (Improved)",
	})
}

// CASLoginSubmitHandlerImproved handles CAS login form submission
func CASLoginSubmitHandlerImproved(c *gin.Context) {
	var req struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
		Service  string `form:"service"`
		Gateway  bool   `form:"gateway"`
		Renew    bool   `form:"renew"`
	}

	if err := c.ShouldBind(&req); err != nil {
		logger.Error("Invalid CAS login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters",
		})
		return
	}

	logger.Info("CAS login submit (improved)",
		zap.String("username", req.Username),
		zap.String("service", req.Service),
		zap.String("ip", c.ClientIP()),
	)

	// Authenticate user using existing logic from auth.go
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		logger.Error("User not found", zap.String("username", req.Username))
		c.HTML(http.StatusUnauthorized, "cas_login.html", gin.H{
			"error":   "Invalid username or password",
			"service": req.Service,
			"gateway": req.Gateway,
			"renew":   req.Renew,
			"title":   "CAS Login (Improved)",
		})
		return
	}

	// Check if user is active
	if user.Status != 1 {
		logger.Error("User is not active", zap.String("username", req.Username))
		c.HTML(http.StatusUnauthorized, "cas_login.html", gin.H{
			"error":   "Account is not active",
			"service": req.Service,
			"gateway": req.Gateway,
			"renew":   req.Renew,
			"title":   "CAS Login (Improved)",
		})
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		logger.Error("User account is locked", zap.String("username", req.Username))
		c.HTML(http.StatusUnauthorized, "cas_login.html", gin.H{
			"error":   "Account is temporarily locked",
			"service": req.Service,
			"gateway": req.Gateway,
			"renew":   req.Renew,
			"title":   "CAS Login (Improved)",
		})
		return
	}

	// Validate password
	if !utils.CheckPassword(req.Password, user.Password) {
		// Update failed login count
		user.FailedCount++
		if user.FailedCount >= 5 {
			// Lock account for 30 minutes
			lockUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockUntil
		}
		database.DB.Save(&user)

		logger.Error("Invalid password", zap.String("username", req.Username))
		c.HTML(http.StatusUnauthorized, "cas_login.html", gin.H{
			"error":   "Invalid username or password",
			"service": req.Service,
			"gateway": req.Gateway,
			"renew":   req.Renew,
			"title":   "CAS Login (Improved)",
		})
		return
	}

	// Reset failed count on successful login
	user.FailedCount = 0
	user.LockedUntil = nil
	now := time.Now()
	user.LastLoginAt = &now
	user.LoginCount++
	database.DB.Save(&user)

	// Create session
	sessionManager := GetSessionManager()
	ctx := context.Background()
	sessionID, err := sessionManager.CreateSession(
		ctx,
		user.ID,
		user.Username,
		user.Email,
		user.DisplayName,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"", // no token ID for CAS
		time.Hour*24,
	)
	if err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create session",
		})
		return
	}

	// Set session cookie
	c.SetCookie("cas_session", sessionID, 86400, "/", "", false, true)

	// Generate service ticket if service is specified
	if req.Service != "" {
		ticket, err := casTicketManager.GenerateServiceTicket(user.Username, req.Service)
		if err != nil {
			logger.Error("Failed to generate CAS service ticket", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate service ticket",
			})
			return
		}

		// Redirect back to service with ticket
		redirectURL := buildCASRedirectURL(req.Service, ticket)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// No service specified, show success page
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user.Username,
	})
}

// CASValidateHandlerImproved handles CAS 1.0 ticket validation
func CASValidateHandlerImproved(c *gin.Context) {
	service := c.Query("service")
	ticket := c.Query("ticket")

	logger.Info("CAS validate request (improved)",
		zap.String("service", service),
		zap.String("ticket", ticket),
		zap.String("ip", c.ClientIP()),
	)

	if service == "" || ticket == "" {
		c.String(http.StatusOK, "no\n")
		return
	}

	// Validate ticket
	username, valid := casTicketManager.ValidateServiceTicket(ticket, service)
	if valid {
		c.String(http.StatusOK, "yes\n%s\n", username)
	} else {
		c.String(http.StatusOK, "no\n")
	}
}

// CASServiceValidateHandlerImproved handles CAS 2.0/3.0 service validation
func CASServiceValidateHandlerImproved(c *gin.Context) {
	service := c.Query("service")
	ticket := c.Query("ticket")
	format := c.Query("format") // xml, json

	logger.Info("CAS service validate request (improved)",
		zap.String("service", service),
		zap.String("ticket", ticket),
		zap.String("format", format),
		zap.String("ip", c.ClientIP()),
	)

	if service == "" || ticket == "" {
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{
				"serviceResponse": gin.H{
					"authenticationFailure": gin.H{
						"code":        "INVALID_REQUEST",
						"description": "Missing required parameters",
					},
				},
			})
		} else {
			c.XML(http.StatusOK, gin.H{
				"cas:serviceResponse": gin.H{
					"cas:authenticationFailure": gin.H{
						"code":     "INVALID_REQUEST",
						"#content": "Missing required parameters",
					},
				},
			})
		}
		return
	}

	// Validate ticket
	username, valid := casTicketManager.ValidateServiceTicket(ticket, service)
	if valid {
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{
				"serviceResponse": gin.H{
					"authenticationSuccess": gin.H{
						"user": username,
						"attributes": gin.H{
							"username": username,
						},
					},
				},
			})
		} else {
			c.XML(http.StatusOK, gin.H{
				"cas:serviceResponse": gin.H{
					"cas:authenticationSuccess": gin.H{
						"cas:user": username,
						"cas:attributes": gin.H{
							"cas:username": username,
						},
					},
				},
			})
		}
	} else {
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{
				"serviceResponse": gin.H{
					"authenticationFailure": gin.H{
						"code":        "INVALID_TICKET",
						"description": "Ticket validation failed",
					},
				},
			})
		} else {
			c.XML(http.StatusOK, gin.H{
				"cas:serviceResponse": gin.H{
					"cas:authenticationFailure": gin.H{
						"code":     "INVALID_TICKET",
						"#content": "Ticket validation failed",
					},
				},
			})
		}
	}
}

// Helper functions
func checkUserAuthenticationFromSession(c *gin.Context) (*models.User, bool) {
	// Check for session cookie
	sessionID, err := c.Cookie("cas_session")
	if err != nil {
		return nil, false
	}

	sessionManager := GetSessionManager()
	ctx := context.Background()
	sessionInfo, err := sessionManager.GetSession(ctx, sessionID)
	if err != nil {
		return nil, false
	}

	var user models.User
	if err := database.DB.Where("id = ?", sessionInfo.UserID).First(&user).Error; err != nil {
		return nil, false
	}

	return &user, true
}

func buildCASRedirectURL(service, ticket string) string {
	u, err := url.Parse(service)
	if err != nil {
		return service + "?ticket=" + ticket
	}

	query := u.Query()
	query.Set("ticket", ticket)
	u.RawQuery = query.Encode()
	return u.String()
}
