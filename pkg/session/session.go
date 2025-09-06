package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"eiam-platform/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID         string    `json:"session_id"`
	UserID            string    `json:"user_id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	DisplayName       string    `json:"display_name"`
	LoginIP           string    `json:"login_ip"`
	UserAgent         string    `json:"user_agent"`
	DeviceFingerprint string    `json:"device_fingerprint"` // 设备指纹
	DeviceType        string    `json:"device_type"`        // 设备类型
	LoginTime         time.Time `json:"login_time"`
	LastActivity      time.Time `json:"last_activity"`
	ExpiresAt         time.Time `json:"expires_at"`
	TokenID           string    `json:"token_id"` // JWT Token ID
}

// SessionManager 会话管理器
type SessionManager struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewSessionManager 创建会话管理器
func NewSessionManager(redisClient *redis.Client, logger *zap.Logger) *SessionManager {
	return &SessionManager{
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(ctx context.Context, userID, username, email, displayName, loginIP, userAgent, tokenID string, expireDuration time.Duration) (string, error) {
	sessionID := utils.GenerateTradeIDString("session")
	now := time.Now()

	// 生成设备指纹
	deviceFingerprint := sm.generateDeviceFingerprint(loginIP, userAgent)
	deviceType := sm.detectDeviceType(userAgent)

	sessionInfo := &SessionInfo{
		SessionID:         sessionID,
		UserID:            userID,
		Username:          username,
		Email:             email,
		DisplayName:       displayName,
		LoginIP:           loginIP,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
		DeviceType:        deviceType,
		LoginTime:         now,
		LastActivity:      now,
		ExpiresAt:         now.Add(expireDuration),
		TokenID:           tokenID,
	}

	// 序列化会话信息
	sessionData, err := json.Marshal(sessionInfo)
	if err != nil {
		sm.logger.Error("Failed to marshal session info", zap.Error(err))
		return "", err
	}

	// 存储到Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionKey := fmt.Sprintf("user_sessions:%s", userID)

	// 获取多设备登录配置
	allowMultiDevice := sm.getMultiDeviceLoginConfig(ctx)

	// 使用Pipeline执行多个操作
	pipe := sm.redisClient.Pipeline()

	// 存储会话信息
	pipe.Set(ctx, sessionKey, sessionData, expireDuration)

	if allowMultiDevice {
		// 多设备模式：添加新会话到列表
		pipe.SAdd(ctx, userSessionKey, sessionID)
		pipe.Expire(ctx, userSessionKey, expireDuration)

		sm.logger.Info("Multi-device session created",
			zap.String("session_id", sessionID),
			zap.String("user_id", userID),
			zap.String("username", username),
			zap.String("device_type", deviceType),
			zap.String("device_fingerprint", deviceFingerprint),
		)
	} else {
		// 单设备模式：清理旧会话，只保留新会话
		pipe.Del(ctx, userSessionKey)
		pipe.SAdd(ctx, userSessionKey, sessionID)
		pipe.Expire(ctx, userSessionKey, expireDuration)

		sm.logger.Info("Single-device session created, old sessions cleaned",
			zap.String("session_id", sessionID),
			zap.String("user_id", userID),
			zap.String("username", username),
			zap.String("device_type", deviceType),
		)
	}

	// 执行Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		sm.logger.Error("Failed to create session in Redis", zap.Error(err))
		return "", err
	}

	return sessionID, nil
}

// GetSession 获取会话信息
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	sessionData, err := sm.redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		sm.logger.Error("Failed to get session from Redis", zap.Error(err))
		return nil, err
	}

	var sessionInfo SessionInfo
	err = json.Unmarshal([]byte(sessionData), &sessionInfo)
	if err != nil {
		sm.logger.Error("Failed to unmarshal session info", zap.Error(err))
		return nil, err
	}

	// 检查会话是否过期
	if time.Now().After(sessionInfo.ExpiresAt) {
		sm.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return &sessionInfo, nil
}

// UpdateActivity 更新会话活动时间
func (sm *SessionManager) UpdateActivity(ctx context.Context, sessionID string) error {
	sessionInfo, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	sessionInfo.LastActivity = time.Now()

	// 序列化更新后的会话信息
	sessionData, err := json.Marshal(sessionInfo)
	if err != nil {
		sm.logger.Error("Failed to marshal updated session info", zap.Error(err))
		return err
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// 更新Redis中的会话信息
	err = sm.redisClient.Set(ctx, sessionKey, sessionData, time.Until(sessionInfo.ExpiresAt)).Err()
	if err != nil {
		sm.logger.Error("Failed to update session activity", zap.Error(err))
		return err
	}

	return nil
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	// 先获取会话信息，以便从用户会话列表中移除
	sessionInfo, err := sm.GetSession(ctx, sessionID)
	if err != nil && err.Error() != "session not found" {
		return err
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// 使用Pipeline执行多个操作
	pipe := sm.redisClient.Pipeline()

	// 删除会话
	pipe.Del(ctx, sessionKey)

	// 从用户会话列表中移除
	if sessionInfo != nil {
		userSessionKey := fmt.Sprintf("user_sessions:%s", sessionInfo.UserID)
		pipe.SRem(ctx, userSessionKey, sessionID)
	}

	// 执行Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		sm.logger.Error("Failed to delete session from Redis", zap.Error(err))
		return err
	}

	sm.logger.Info("Session deleted", zap.String("session_id", sessionID))
	return nil
}

// GetUserSessions 获取用户的所有会话
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*SessionInfo, error) {
	userSessionKey := fmt.Sprintf("user_sessions:%s", userID)

	sessionIDs, err := sm.redisClient.SMembers(ctx, userSessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*SessionInfo{}, nil
		}
		sm.logger.Error("Failed to get user sessions from Redis", zap.Error(err))
		return nil, err
	}

	var sessions []*SessionInfo
	var expiredSessionIDs []string

	// 检查每个会话并收集过期的会话ID
	for _, sessionID := range sessionIDs {
		sessionInfo, err := sm.GetSession(ctx, sessionID)
		if err != nil {
			// 如果会话已过期或不存在，收集起来稍后批量移除
			expiredSessionIDs = append(expiredSessionIDs, sessionID)
			continue
		}
		sessions = append(sessions, sessionInfo)
	}

	// 批量移除过期的会话ID
	if len(expiredSessionIDs) > 0 {
		pipe := sm.redisClient.Pipeline()
		for _, sessionID := range expiredSessionIDs {
			pipe.SRem(ctx, userSessionKey, sessionID)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			sm.logger.Error("Failed to remove expired session IDs from user session list", zap.Error(err))
		} else {
			sm.logger.Info("Cleaned up expired session IDs",
				zap.String("user_id", userID),
				zap.Int("expired_count", len(expiredSessionIDs)),
			)
		}
	}

	return sessions, nil
}

// ForceLogoutUser 强制用户下线（踢人）
func (sm *SessionManager) ForceLogoutUser(ctx context.Context, userID string) error {
	sessions, err := sm.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// 删除用户的所有会话
	for _, session := range sessions {
		err := sm.DeleteSession(ctx, session.SessionID)
		if err != nil {
			sm.logger.Error("Failed to delete session during force logout",
				zap.String("session_id", session.SessionID),
				zap.Error(err),
			)
		}
	}

	sm.logger.Info("User forced logout",
		zap.String("user_id", userID),
		zap.Int("sessions_count", len(sessions)),
	)

	return nil
}

// ForceLogoutSession 强制特定会话下线
func (sm *SessionManager) ForceLogoutSession(ctx context.Context, sessionID string) error {
	return sm.DeleteSession(ctx, sessionID)
}

// CleanExpiredSessions 清理过期会话（定时任务）
func (sm *SessionManager) CleanExpiredSessions(ctx context.Context) error {
	// 这个方法可以作为定时任务运行，清理Redis中的过期会话
	// 由于Redis会自动过期键，这里主要是清理用户会话列表中的无效引用

	// 获取所有用户会话列表
	pattern := "user_sessions:*"
	keys, err := sm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, userSessionKey := range keys {
		sessionIDs, err := sm.redisClient.SMembers(ctx, userSessionKey).Result()
		if err != nil {
			continue
		}

		for _, sessionID := range sessionIDs {
			sessionKey := fmt.Sprintf("session:%s", sessionID)
			exists, err := sm.redisClient.Exists(ctx, sessionKey).Result()
			if err != nil || exists == 0 {
				// 会话不存在，从用户会话列表中移除
				sm.redisClient.SRem(ctx, userSessionKey, sessionID)
			}
		}
	}

	return nil
}

// GetActiveSessionsCount 获取活跃会话数量
func (sm *SessionManager) GetActiveSessionsCount(ctx context.Context) (int64, error) {
	pattern := "session:*"
	keys, err := sm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

// IsSessionValid 检查会话是否有效
func (sm *SessionManager) IsSessionValid(ctx context.Context, sessionID string) bool {
	_, err := sm.GetSession(ctx, sessionID)
	return err == nil
}

// BlacklistToken 将token加入黑名单
func (sm *SessionManager) BlacklistToken(ctx context.Context, tokenID string, expireDuration time.Duration) error {
	blacklistKey := fmt.Sprintf("token_blacklist:%s", tokenID)
	err := sm.redisClient.Set(ctx, blacklistKey, "revoked", expireDuration).Err()
	if err != nil {
		sm.logger.Error("Failed to blacklist token", zap.Error(err), zap.String("token_id", tokenID))
		return err
	}

	sm.logger.Info("Token blacklisted", zap.String("token_id", tokenID))
	return nil
}

// IsTokenBlacklisted 检查token是否在黑名单中
func (sm *SessionManager) IsTokenBlacklisted(ctx context.Context, tokenID string) bool {
	blacklistKey := fmt.Sprintf("token_blacklist:%s", tokenID)
	exists, err := sm.redisClient.Exists(ctx, blacklistKey).Result()
	if err != nil {
		sm.logger.Error("Failed to check token blacklist", zap.Error(err), zap.String("token_id", tokenID))
		return false
	}
	return exists > 0
}

// CleanExpiredBlacklistedTokens 清理过期的黑名单token
func (sm *SessionManager) CleanExpiredBlacklistedTokens(ctx context.Context) error {
	// Redis会自动清理过期的key，这里主要是记录日志
	pattern := "token_blacklist:*"
	keys, err := sm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	sm.logger.Info("Blacklisted tokens cleanup", zap.Int("count", len(keys)))
	return nil
}

// generateDeviceFingerprint 生成设备指纹
func (sm *SessionManager) generateDeviceFingerprint(ip, userAgent string) string {
	// 简单的设备指纹生成：IP + User-Agent 的哈希
	fingerprint := ip + "|" + userAgent
	return utils.GenerateTradeIDString("device") + "_" + utils.GenerateTradeIDString(fingerprint)[:8]
}

// detectDeviceType 检测设备类型
func (sm *SessionManager) detectDeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipipad") {
		return "Mobile"
	} else if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		return "Tablet"
	} else if strings.Contains(userAgent, "windows") || strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "linux") {
		return "Desktop"
	} else if strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "crawler") {
		return "Bot"
	}

	return "Unknown"
}

// getMultiDeviceLoginConfig 获取多设备登录配置
func (sm *SessionManager) getMultiDeviceLoginConfig(ctx context.Context) bool {
	// 从Redis获取配置，如果没有配置则默认为false（单设备模式）
	configKey := "system:security:allow_multi_device_login"
	value, err := sm.redisClient.Get(ctx, configKey).Result()
	if err != nil {
		// 如果没有配置，返回默认值false
		return false
	}

	// 解析配置值
	allowMultiDevice, _ := strconv.ParseBool(value)
	return allowMultiDevice
}

// ForceLogoutAllUsers 强制所有用户下线
func (sm *SessionManager) ForceLogoutAllUsers(ctx context.Context) (int, error) {
	// 获取所有会话key
	pattern := "session:*"
	keys, err := sm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		sm.logger.Error("Failed to get session keys", zap.Error(err))
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	// 获取所有用户会话列表key
	userSessionPattern := "user_sessions:*"
	userSessionKeys, err := sm.redisClient.Keys(ctx, userSessionPattern).Result()
	if err != nil {
		sm.logger.Error("Failed to get user session keys", zap.Error(err))
		return 0, err
	}

	// 使用Pipeline批量删除
	pipe := sm.redisClient.Pipeline()

	// 删除所有会话
	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	// 删除所有用户会话列表
	for _, key := range userSessionKeys {
		pipe.Del(ctx, key)
	}

	// 执行Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		sm.logger.Error("Failed to execute pipeline for force logout all users", zap.Error(err))
		return 0, err
	}

	affectedCount := len(keys) + len(userSessionKeys)
	sm.logger.Info("All users force logged out",
		zap.Int("session_keys_deleted", len(keys)),
		zap.Int("user_session_keys_deleted", len(userSessionKeys)),
		zap.Int("total_affected", affectedCount),
	)

	return affectedCount, nil
}
