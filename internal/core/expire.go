package core

import (
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
)

func (e *Executor) ActiveDeleteExpiredKeys() {
	for {
		var expiredCount = 0
		var sampleCountRemain = constant.ActiveExpireSampleSize
		for key, expiredTime := range e.store.Dict.GetExpireDictStore() {
			sampleCountRemain--
			if sampleCountRemain < 0 {
				break
			}
			if time.Now().UnixMilli() > int64(expiredTime) {
				e.store.Dict.Delete(key)
				expiredCount++
			}
		}

		if float64(expiredCount)/float64(constant.ActiveExpireSampleSize) <= constant.ActiveExpireThreshold {
			break
		}
	}
}
