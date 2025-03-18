package ntpsync

import (
	"time"
)

// NTP时间戳是相对于1900-01-01T00:00:00Z的
const ntpEpoch = 2208988800

// timeToNTPTime 将time.Time转换为NTP秒和小数部分
func timeToNTPTime(t time.Time) (uint32, uint32) {
	seconds := uint32(t.Unix() + ntpEpoch)
	fraction := uint32(t.Nanosecond() * 0x100000000 / 1000000000)
	return seconds, fraction
}

// ntpTimeToTime 将NTP秒和小数部分转换为time.Time
func ntpTimeToTime(seconds, fraction uint32) time.Time {
	secs := int64(seconds - ntpEpoch)
	nanos := int64(fraction) * 1000000000 / 0x100000000
	return time.Unix(secs, nanos)
}
