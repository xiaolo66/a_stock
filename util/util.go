package util

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bensema/gotdx/types"
)

// ParseMarket 将 SZ/SH/BJ（大小写均可）映射为 gotdx 市场枚举。
func ParseMarket(market string) (types.Market, error) {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case types.MarketSZName, "sz":
		return types.MarketSZ, nil
	case types.MarketSHName, "sh":
		return types.MarketSH, nil
	case types.MarketBJName, "bj":
		return types.MarketBJ, nil
	default:
		return 0, fmt.Errorf("unknown market %q, want SZ, SH or BJ", market)
	}
}

// MarketsForQuery 空字符串返回沪深北三个市场，否则解析单个市场。
func MarketsForQuery(market string) ([]types.Market, error) {
	if strings.TrimSpace(market) == "" {
		return []types.Market{types.MarketSZ, types.MarketSH}, nil
	}
	m, err := ParseMarket(market)
	if err != nil {
		return nil, err
	}
	return []types.Market{m}, nil
}

// ParseCategory 将分类字符串映射为 gotdx category；空字符串默认 A 股。
func ParseCategory(category string) (uint8, error) {
	switch strings.ToUpper(strings.TrimSpace(category)) {
	case "", "A":
		return types.CategoryA, nil
	case "SH":
		return types.CategorySH, nil
	case "SZ":
		return types.CategorySZ, nil
	case "BJ":
		return types.CategoryBJ, nil
	default:
		return 0, fmt.Errorf("unknown category %q, want A, SH, SZ, KCB, CYB or BJ", category)
	}
}

// ParseKLineCategory 将 K 线周期字符串映射为 gotdx category。
// 5m→0, 15m→1, 30m→2, 1h→3, 1d→4, 1w→5, 1mo→6, ex1m→7, 1m→8, rik→9, 1q→10, 1y→11。
func ParseKLineCategory(category string) (uint16, error) {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "5m":
		return types.KLINE_TYPE_5MIN, nil
	case "15m":
		return types.KLINE_TYPE_15MIN, nil
	case "30m":
		return types.KLINE_TYPE_30MIN, nil
	case "1h", "60m":
		return types.KLINE_TYPE_1HOUR, nil
	case "1d":
		return types.KLINE_TYPE_DAILY, nil
	case "1w":
		return types.KLINE_TYPE_WEEKLY, nil
	case "1mo":
		return types.KLINE_TYPE_MONTHLY, nil
	case "ex1m":
		return types.KLINE_TYPE_EXHQ_1MIN, nil
	case "1m":
		return types.KLINE_TYPE_1MIN, nil
	case "rik":
		return types.KLINE_TYPE_RI_K, nil
	case "1q", "3mo":
		return types.KLINE_TYPE_3MONTH, nil
	case "1y":
		return types.KLINE_TYPE_YEARLY, nil
	default:
		return 0, fmt.Errorf("unknown kline category %q, want 5m, 15m, 30m, 1h, 1d, 1w, 1mo, ex1m, 1m, rik, 1q, 1y", category)
	}
}

func IsAStockCode(code string, market types.Market) bool {
	code = types.CleanCode(code)
	if len(code) != 6 {
		return false
	}
	switch market {
	case types.MarketSH: // 沪市
		switch code[:3] {
		case "600", "601", "603", "605", "688":
			return true
		default:
			return false
		}

	case types.MarketSZ: // 深市
		switch code[:3] {
		case "000", "001", "002", "003", "300", "301":
			return true
		default:
			return false
		}

	case types.MarketBJ: // 北交所：前两位 43、83、87、92
		switch code[:2] {
		case "43", "83", "87", "92":
			return true
		default:
			return false
		}

	default:
		return false
	}
}

func CalcChangeAndRate(price, preClose float64) (change, changeRate float64) {
	change = round2(price - preClose)
	if preClose == 0 {
		return change, 0
	}
	changeRate = round2(change / preClose * 100)
	return change, changeRate
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Round2 保留两位小数。
func Round2(v float64) float64 {
	return round2(v)
}

// MinuteTimeAtIndex 将分时序号映射为 A 股交易时间（跳过 11:30-13:00 午休）。
func MinuteTimeAtIndex(index int) string {
	var totalMin int
	if index < 120 {
		totalMin = 9*60 + 30 + index
	} else {
		totalMin = 13*60 + (index - 120)
	}
	return fmt.Sprintf("%02d:%02d", totalMin/60, totalMin%60)
}

// ParseDate 解析 YYYY-MM-DD 或 YYYYMMDD。
func ParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("date is empty")
	}
	layouts := []string{"2006-01-02", "20060102"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return DateOnly(t), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q, want YYYY-MM-DD or YYYYMMDD", s)
}

// DateOnly 取本地日期的零点。
func DateOnly(t time.Time) time.Time {
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// TradingDaysInclusive 统计 [from, to] 内的工作日数量（不含法定节假日）。
func TradingDaysInclusive(from, to time.Time) int {
	from = DateOnly(from)
	to = DateOnly(to)
	if from.After(to) {
		return 0
	}
	n := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			n++
		}
	}
	return n
}

// EstimateKLineCountSince 估算从 from 到 to 的 K 线根数，用于减少分页请求。
func EstimateKLineCountSince(from, to time.Time, category uint16) int {
	if to.Before(from) {
		return 0
	}
	switch category {
	case types.KLINE_TYPE_DAILY, types.KLINE_TYPE_RI_K:
		return TradingDaysInclusive(from, to)
	case types.KLINE_TYPE_WEEKLY:
		days := int(DateOnly(to).Sub(DateOnly(from)).Hours()/24) + 1
		return (days + 6) / 7
	case types.KLINE_TYPE_MONTHLY:
		y1, m1, _ := from.Date()
		y2, m2, _ := to.Date()
		return (y2-y1)*12 + int(m2-m1) + 1
	case types.KLINE_TYPE_3MONTH:
		months := EstimateKLineCountSince(from, to, types.KLINE_TYPE_MONTHLY)
		return (months + 2) / 3
	case types.KLINE_TYPE_YEARLY:
		return to.Year() - from.Year() + 1
	default:
		days := TradingDaysInclusive(from, to)
		if days == 0 {
			return 0
		}
		barsPerDay := 240
		switch category {
		case types.KLINE_TYPE_1MIN, types.KLINE_TYPE_EXHQ_1MIN:
			barsPerDay = 240
		case types.KLINE_TYPE_5MIN:
			barsPerDay = 48
		case types.KLINE_TYPE_15MIN:
			barsPerDay = 16
		case types.KLINE_TYPE_30MIN:
			barsPerDay = 8
		case types.KLINE_TYPE_1HOUR:
			barsPerDay = 4
		}
		return days * barsPerDay
	}
}

// FormatKLineDateTime 按 K 线周期格式化时间。
func FormatKLineDateTime(dt time.Time, category uint16) string {
	switch category {
	case types.KLINE_TYPE_DAILY, types.KLINE_TYPE_RI_K,
		types.KLINE_TYPE_WEEKLY, types.KLINE_TYPE_MONTHLY,
		types.KLINE_TYPE_3MONTH, types.KLINE_TYPE_YEARLY:
		return dt.Format("2006-01-02")
	default:
		return dt.Format("2006-01-02 15:04")
	}
}
