package ratelimit

import (
	"feedSystem_video/internal/apierror"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens   float64   // 当前令牌数
	lastTime time.Time // 上次补充时间
	mu       sync.Mutex
}

// Limiter 基于客户端 IP 的令牌桶限流器
type Limiter struct {
	buckets    map[string]*tokenBucket // 每个 IP 一个桶
	mu         sync.RWMutex
	rate       float64 // 每秒生成的令牌数
	burst      float64 // 桶容量（最大突发请求数）
	cleanupSec int     // 清理间隔（秒）
}

// NewLimiter 创建限流器
// rate: 每秒允许的请求数，burst: 突发上限
func NewLimiter(rate, burst float64) *Limiter {
	l := &Limiter{
		buckets:    make(map[string]*tokenBucket),
		rate:       rate,
		burst:      burst,
		cleanupSec: 60,
	}
	// 启动后台清理 goroutine，定期清除过期的桶避免内存泄漏
	go l.cleanup()
	return l
}

// Allow 检查指定 key（通常是 IP）是否允许通过
func (l *Limiter) Allow(key string) bool {
	l.mu.RLock()
	bucket, exists := l.buckets[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		// 双重检查，防止并发创建多个桶
		bucket, exists = l.buckets[key]
		if !exists {
			bucket = &tokenBucket{
				tokens:   l.burst,
				lastTime: time.Now(),
			}
			l.buckets[key] = bucket
		}
		l.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	// 按时间差补充令牌
	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens += elapsed * l.rate
	if bucket.tokens > l.burst {
		bucket.tokens = l.burst
	}
	bucket.lastTime = now

	if bucket.tokens >= 1 {
		bucket.tokens -= 1
		return true
	}
	return false
}

// cleanup 定期清理超过 5 分钟未使用的桶
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(time.Duration(l.cleanupSec) * time.Second)
	for range ticker.C {
		l.mu.Lock()
		for ip, bucket := range l.buckets {
			bucket.mu.Lock()
			if time.Since(bucket.lastTime) > 5*time.Minute {
				delete(l.buckets, ip)
			}
			bucket.mu.Unlock()
		}
		l.mu.Unlock()
	}
}

// Middleware 返回 Gin 限流中间件
// 默认每秒 10 个请求，突发上限 20
func Middleware(rate, burst float64) gin.HandlerFunc {
	limiter := NewLimiter(rate, burst)
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			apierror.Fail(c, 429, 4290, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
