package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/artem/wg-orbit/internal/auth"
	"github.com/artem/wg-orbit/internal/storage"
	"github.com/artem/wg-orbit/internal/wg"
)

// Server представляє REST API сервер
type Server struct {
	storage      storage.Storage
	tokenManager *auth.TokenManager
	config       *Config
}

// Config конфігурація для REST API
type Config struct {
	Port      int    `yaml:"port" json:"port"`
	Host      string `yaml:"host" json:"host"`
	TLSCert   string `yaml:"tls_cert" json:"tls_cert"`
	TLSKey    string `yaml:"tls_key" json:"tls_key"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
}

// NewServer створює новий REST API сервер
func NewServer(storage storage.Storage, tokenManager *auth.TokenManager, config *Config) *Server {
	return &Server{
		storage:      storage,
		tokenManager: tokenManager,
		config:       config,
	}
}

// SetupRoutes налаштовує маршрути API
func (s *Server) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Middleware для CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Публічні маршрути
	public := r.Group("/api/v1")
	{
		public.POST("/enroll", s.handleEnroll)
		public.GET("/health", s.handleHealth)
	}

	// Захищені маршрути
	protected := r.Group("/api/v1")
	protected.Use(s.authMiddleware())
	{
		// Peer management
		protected.GET("/peers", s.handleListPeers)
		protected.GET("/peers/:id", s.handleGetPeer)
		protected.POST("/peers", s.handleCreatePeer)
		protected.PUT("/peers/:id", s.handleUpdatePeer)
		protected.DELETE("/peers/:id", s.handleDeletePeer)

		// Configuration
		protected.GET("/config/:peer_id", s.handleGetConfig)
		protected.POST("/refresh-token", s.handleRefreshToken)
	}

	return r
}

// authMiddleware перевіряє JWT токен
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := s.tokenManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// handleHealth перевіряє стан сервера
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// handleEnroll обробляє реєстрацію клієнта
func (s *Server) handleEnroll(c *gin.Context) {
	var req struct {
		Token      string `json:"token" binding:"required"`
		PublicKey  string `json:"public_key" binding:"required"`
		ClientName string `json:"client_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валідуємо enrollment token
	claims, err := s.tokenManager.ValidateToken(req.Token)
	if err != nil || claims.Role != "enrollment" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid enrollment token"})
		return
	}

	// Перевіряємо, чи не існує вже peer з таким ім'ям
	existingPeer, err := s.storage.GetPeerByName(req.ClientName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if existingPeer != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Peer with this name already exists"})
		return
	}

	// Створюємо нового peer'а
	peer := &wg.Peer{
		ID:         uuid.New(),
		Name:       req.ClientName,
		PublicKey:  req.PublicKey,
		AllowedIPs: []string{"10.0.0.0/24"}, // TODO: Implement proper IPAM
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}

	if err := s.storage.SavePeer(peer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save peer"})
		return
	}

	// Генеруємо постійний токен для клієнта
	accessToken, err := s.tokenManager.GenerateToken(
		claims.UserID, req.ClientName, "client", &peer.ID, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"peer_id":      peer.ID,
		"access_token": accessToken,
		"allowed_ips":  peer.AllowedIPs,
	})
}

// handleListPeers повертає список всіх peer'ів
func (s *Server) handleListPeers(c *gin.Context) {
	peers, err := s.storage.ListPeers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch peers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"peers": peers})
}

// handleGetPeer повертає інформацію про конкретний peer
func (s *Server) handleGetPeer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer ID"})
		return
	}

	peer, err := s.storage.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if peer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	c.JSON(http.StatusOK, peer)
}

// handleCreatePeer створює новий peer
func (s *Server) handleCreatePeer(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peer, err := wg.NewPeer(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create peer"})
		return
	}

	// TODO: Implement proper IPAM
	peer.AllowedIPs = []string{"10.0.0.0/32"}

	if err := s.storage.SavePeer(peer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save peer"})
		return
	}

	c.JSON(http.StatusCreated, peer)
}

// handleUpdatePeer оновлює peer
func (s *Server) handleUpdatePeer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer ID"})
		return
	}

	peer, err := s.storage.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if peer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	var req struct {
		Name      *string  `json:"name"`
		IsActive  *bool    `json:"is_active"`
		Endpoint  *string  `json:"endpoint"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		peer.Name = *req.Name
	}
	if req.IsActive != nil {
		peer.IsActive = *req.IsActive
	}
	if req.Endpoint != nil {
		peer.Endpoint = *req.Endpoint
	}

	peer.UpdatedAt = time.Now()

	if err := s.storage.SavePeer(peer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update peer"})
		return
	}

	c.JSON(http.StatusOK, peer)
}

// handleDeletePeer видаляє peer
func (s *Server) handleDeletePeer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer ID"})
		return
	}

	if err := s.storage.DeletePeer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete peer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peer deleted successfully"})
}

// handleGetConfig повертає WireGuard конфігурацію для клієнта
func (s *Server) handleGetConfig(c *gin.Context) {
	peerIDStr := c.Param("peer_id")
	peerID, err := uuid.Parse(peerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer ID"})
		return
	}

	peer, err := s.storage.GetPeer(peerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if peer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	// Отримуємо інформацію про сервер
	serverInterface, err := s.storage.GetInterface("wg0")
	if err != nil || serverInterface == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server interface not configured"})
		return
	}

	// Створюємо конфігурацію для клієнта
	clientConfig := &wg.ClientConfig{
		Interface: wg.ClientInterface{
			PrivateKey: peer.PrivateKey,
			Address:    peer.AllowedIPs,
			DNS:        []string{"8.8.8.8", "8.8.4.4"},
		},
		Peer: wg.ServerPeer{
			PublicKey:  serverInterface.PublicKey,
			Endpoint:   s.config.Host + ":" + strconv.Itoa(serverInterface.ListenPort),
			AllowedIPs: []string{"0.0.0.0/0"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"config":     clientConfig,
		"config_wg": clientConfig.ToWireGuardConfig(),
	})
}

// handleRefreshToken оновлює токен
func (s *Server) handleRefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")

	// Генеруємо новий токен
	newToken, err := s.tokenManager.GenerateToken(
		userID.(uuid.UUID), username.(string), role.(string), nil, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newToken})
}

// Start запускає сервер
func (s *Server) Start() error {
	r := s.SetupRoutes()

	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)

	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		return r.RunTLS(addr, s.config.TLSCert, s.config.TLSKey)
	}

	return r.Run(addr)
}