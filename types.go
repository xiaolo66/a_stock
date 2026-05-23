package a_stock

type StockCode struct {
	Name   string // 证券简称（列表接口返回时填充）
	Code   string // 六位代码，如 000001
	Market string // SZ / SH
	Symbol string // 带交易所后缀，如 000001.SZ
}

type QuoteLevel struct {
	Price float64 // 档位价格
	Vol   int     // 档位挂单量
}

// StockQuote 证券五档快照行情。
type StockQuote struct {
	Code       string       // 六位代码
	Market     string       // SZ / SH / BJ
	Symbol     string       // 如 000001.SZ
	Price      float64      // 最新价
	PreClose   float64      // 昨收
	Change     float64      // 涨跌额 = 最新价 - 昨收
	ChangeRate float64      // 涨跌幅(%) = (最新价 - 昨收) / 昨收 * 100
	Open       float64      // 今开
	High       float64      // 最高
	Low        float64      // 最低
	Vol        int          // 成交量
	Amount     float64      // 成交额
	Turnover   float64      // 换手率
	ServerTime string       // 行情时间
	BidLevels  []QuoteLevel // 买盘五档
	AskLevels  []QuoteLevel // 卖盘五档
}

// 排行榜类型，对应 GetStockTopBoard 的 boardType 参数。
const (
	TopBoardIncrease           = 1 // 涨幅榜
	TopBoardDecrease           = 2 // 跌幅榜
	TopBoardAmplitude          = 3 // 振幅榜
	TopBoardRiseSpeed          = 4 // 涨速榜
	TopBoardFallSpeed          = 5 // 跌速榜
	TopBoardVolRatio           = 6 // 量比榜
	TopBoardPosCommissionRatio = 7 // 委比正榜
	TopBoardNegCommissionRatio = 8 // 委比负榜
	TopBoardTurnover           = 9 // 换手率榜
)

func topBoardName(boardType int) string {
	switch boardType {
	case TopBoardIncrease:
		return "涨幅榜"
	case TopBoardDecrease:
		return "跌幅榜"
	case TopBoardAmplitude:
		return "振幅榜"
	case TopBoardRiseSpeed:
		return "涨速榜"
	case TopBoardFallSpeed:
		return "跌速榜"
	case TopBoardVolRatio:
		return "量比榜"
	case TopBoardPosCommissionRatio:
		return "委比正榜"
	case TopBoardNegCommissionRatio:
		return "委比负榜"
	case TopBoardTurnover:
		return "换手率榜"
	default:
		return ""
	}
}

// TopBoardItem 排行榜条目。
type TopBoardItem struct {
	Code   string  // 六位代码
	Market string  // SZ / SH / BJ
	Symbol string  // 如 000001.SZ
	Price  float64 // 最新价
	Value  float64 // 榜单指标值
}

// TopBoardResult 排行榜结果。
type TopBoardResult struct {
	Type  int            // 榜单类型，见 TopBoard* 常量
	Name  string         // 榜单中文名，如 涨幅榜
	Items []TopBoardItem // 榜单数据
}

// HistoryTickPoint 历史分时单分钟数据。
type HistoryTickPoint struct {
	Time  string  // 时间 HH:MM，首条 09:30
	Price float64 // 最新价
	Avg   float64 // 均价
	Vol   int     // 该分钟成交量(手)
}

// HistoryTickChartResult 历史分时结果。
type HistoryTickChartResult struct {
	Date        uint32             // 交易日期 YYYYMMDD
	Symbol      string             // 如 000001.SZ
	Items       []HistoryTickPoint // 分时明细
	TotalVol    int64              // 当日总成交量(手)
	TotalAmount float64            // 当日总成交额(元)
	AvgPrice    float64            // 当日成交均价 = 总成交额 / 总成交量
}

const klinePageSize = uint16(200)

// StockKLineBar K 线单根数据。
type StockKLineBar struct {
	Last       float64 // 昨收
	Open       float64 // 开盘价
	Close      float64 // 收盘价
	High       float64 // 最高价
	Low        float64 // 最低价
	Vol        float64 // 成交量
	Amount     float64 // 成交额
	Turnover   float64 // 换手率(%)
	Change     float64 // 涨跌额
	ChangeRate float64 // 涨跌幅(%)
	DateTime   string  // 时间，日 K 为 YYYY-MM-DD，分钟/小时 K 为 YYYY-MM-DD HH:MM
}
