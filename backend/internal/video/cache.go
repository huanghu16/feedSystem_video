package video

import "time"

// 热榜缓存常量
// 与首页计数策略相反：热榜是"趋势"数据，允许秒级滞后，
// 用短 TTL 缓存扛住高频访问，避免每次都做全表算分 + 排序
const (
	CacheKeyHot = "v1:video:hot"
	CacheTTLHot = 30 * time.Second
)
