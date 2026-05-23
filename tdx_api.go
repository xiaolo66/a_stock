package a_stock

import (
	"fmt"
	"sort"
	"time"

	"a_stock/util"

	"github.com/bensema/gotdx/proto"
	"github.com/bensema/gotdx/types"
)

// 获取a股市场股票代码
func (c *Client) GetAllAStockCodes(market string) ([]StockCode, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}

	markets, err := util.MarketsForQuery(market)
	if err != nil {
		return nil, err
	}

	var out []StockCode
	for _, m := range markets {
		total, err := c.tdx.StockCount(m.Uint8())
		if err != nil {
			return nil, fmt.Errorf("list market %s: %w", m.String(), err)
		}

		for start := uint32(0); start < uint32(total); start += 1600 {
			page, err := c.tdx.StockList(m.Uint8(), start, 1600)
			if err != nil {
				return nil, fmt.Errorf("list market %s: %w", m.String(), err)
			}
			if len(page) == 0 {
				break
			}
			for _, sec := range page {
				code := types.CleanCode(sec.Code)
				if !util.IsAStockCode(code, m) {
					continue
				}
				out = append(out, StockCode{
					Name:   sec.Name,
					Code:   code,
					Market: m.String(),
					Symbol: fmt.Sprintf("%s.%s", code, m.String()),
				})
			}
			if uint32(len(page)) < 1600 {
				break
			}
		}
	}
	return out, nil
}

// 批量获取证券五档快照深度和当日行情，symbols 如 000001.SZ、600000.SH。
func (c *Client) GetStockQuotesDetail(symbols []string) ([]StockQuote, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	if len(symbols) == 0 {
		return nil, fmt.Errorf("symbols is empty")
	}

	markets := make([]uint8, 0, len(symbols))
	codes := make([]string, 0, len(symbols))
	symbolByKey := make(map[string]string, len(symbols))
	for _, symbol := range symbols {
		market, code, err := types.DetectMarket(symbol)
		if err != nil {
			return nil, fmt.Errorf("parse symbol %q: %w", symbol, err)
		}
		key := fmt.Sprintf("%s.%s", code, market.String())
		markets = append(markets, market.Uint8())
		codes = append(codes, code)
		symbolByKey[key] = fmt.Sprintf("%s.%s", code, market.String())
	}

	items, err := c.tdx.StockQuotesDetail(markets, codes)
	if err != nil {
		return nil, err
	}

	out := make([]StockQuote, 0, len(items))
	for _, item := range items {
		market := types.Market(item.Market).String()
		code := types.CleanCode(item.Code)
		key := fmt.Sprintf("%s.%s", code, market)
		out = append(out, toStockQuote(item, key))
	}
	return out, nil
}

// GetStockHistoryTickChart 获取历史日期的每分钟成交和成交汇总，date 格式 YYYYMMDD，如 20260316。
func (c *Client) GetStockHistoryTickChart(symbol string, date uint32) (*HistoryTickChartResult, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	market, code, err := types.DetectMarket(symbol)
	if err != nil {
		return nil, fmt.Errorf("parse symbol %q: %w", symbol, err)
	}
	items, err := c.tdx.StockHistoryTickChart(date, market.Uint8(), code)
	if err != nil {
		return nil, err
	}

	out := &HistoryTickChartResult{
		Date:   date,
		Symbol: fmt.Sprintf("%s.%s", code, market.String()),
		Items:  make([]HistoryTickPoint, len(items)),
	}
	var totalAmount float64
	for i, item := range items {
		out.Items[i] = HistoryTickPoint{
			Time:  util.MinuteTimeAtIndex(i),
			Price: item.Price,
			Avg:   item.Avg,
			Vol:   item.Vol,
		}
		out.TotalVol += int64(item.Vol)
		totalAmount += item.Avg * float64(item.Vol) * 100
	}
	out.TotalAmount = util.Round2(totalAmount)
	if out.TotalVol > 0 {
		out.AvgPrice = util.Round2(totalAmount / float64(out.TotalVol))
	}
	return out, nil
}

// GetStockTickChart 获取当日每分钟成交，symbol 如 000001.SZ，start/count 为分页偏移。
func (c *Client) GetStockTickChart(symbol string, start, count uint16) (*HistoryTickChartResult, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	market, code, err := types.DetectMarket(symbol)
	if err != nil {
		return nil, fmt.Errorf("parse symbol %q: %w", symbol, err)
	}
	items, err := c.tdx.StockTickChart(market.Uint8(), code, start, count)
	if err != nil {
		return nil, err
	}

	out := &HistoryTickChartResult{
		Symbol: fmt.Sprintf("%s.%s", code, market.String()),
		Items:  make([]HistoryTickPoint, len(items)),
	}
	var totalAmount float64
	for i, item := range items {
		out.Items[i] = HistoryTickPoint{
			Time:  util.MinuteTimeAtIndex(int(start) + i),
			Price: item.Price,
			Avg:   item.Avg,
			Vol:   item.Vol,
		}
		out.TotalVol += int64(item.Vol)
		totalAmount += item.Avg * float64(item.Vol) * 100
	}
	out.TotalAmount = util.Round2(totalAmount)
	if out.TotalVol > 0 {
		out.AvgPrice = util.Round2(totalAmount / float64(out.TotalVol))
	}
	return out, nil
}

// GetStockKLineSince 从 fromDate 至今拉取 K 线，条数按日期自动计算。fromDate 支持 YYYY-MM-DD 或 YYYYMMDD。
func (c *Client) GetStockKLineSince(symbol string, category string, fromDate string) ([]StockKLineBar, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	from, err := util.ParseDate(fromDate)
	if err != nil {
		return nil, err
	}
	cat, err := util.ParseKLineCategory(category)
	if err != nil {
		return nil, err
	}
	market, code, err := types.DetectMarket(symbol)
	if err != nil {
		return nil, fmt.Errorf("parse symbol %q: %w", symbol, err)
	}

	to := util.DateOnly(time.Now())
	if from.After(to) {
		return nil, nil
	}

	var raw []proto.SecurityBar
	start := uint16(0)
	for {
		batch, err := c.tdx.StockKLine(cat, market.Uint8(), code, start, klinePageSize, 1, types.AdjustQFQ)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		reachedBeforeFrom := false
		for _, bar := range batch {
			if bar.DateTime.Before(from) {
				reachedBeforeFrom = true
				continue
			}
			if util.DateOnly(bar.DateTime).After(to) {
				continue
			}
			raw = append(raw, bar)
		}

		if reachedBeforeFrom || len(batch) < int(klinePageSize) {
			break
		}
		start += klinePageSize
	}

	sort.Slice(raw, func(i, j int) bool {
		return raw[i].DateTime.Before(raw[j].DateTime)
	})
	return toStockKLineBars(raw, cat), nil
}

// GetStockTopBoard 获取排行榜，boardType 见 TopBoard* 常量，size 为榜单条数。
func (c *Client) GetStockTopBoard(category string, boardType int, size uint8) (*TopBoardResult, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	cat, err := util.ParseCategory(category)
	if err != nil {
		return nil, err
	}
	if boardType < TopBoardIncrease || boardType > TopBoardTurnover {
		return nil, fmt.Errorf("unknown board type %d, want 1-9", boardType)
	}
	if size == 0 {
		size = types.DefaultTopBoardSize
	}

	reply, err := c.tdx.StockTopBoard(cat, size)
	if err != nil {
		return nil, err
	}
	items := selectTopBoard(reply, boardType)
	return &TopBoardResult{
		Type:  boardType,
		Name:  topBoardName(boardType),
		Items: toTopBoardItems(items),
	}, nil
}

// GetStockUnusual 获取盘中异动，market 如 SZ/SH/BJ，start/count 用于分页。
func (c *Client) GetStockUnusual(market string, start, count uint32) ([]proto.UnusualData, error) {
	if c == nil || c.tdx == nil {
		return nil, fmt.Errorf("stock client is nil")
	}
	m, err := util.ParseMarket(market)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		count = 20
	}
	return c.tdx.StockUnusual(m.Uint8(), start, count)
}

func toStockKLineBars(bars []proto.SecurityBar, category uint16) []StockKLineBar {
	if len(bars) == 0 {
		return nil
	}
	out := make([]StockKLineBar, len(bars))
	for i, bar := range bars {
		out[i] = StockKLineBar{
			Last:       bar.Last,
			Open:       bar.Open,
			Close:      bar.Close,
			High:       bar.High,
			Low:        bar.Low,
			Vol:        bar.Vol,
			Amount:     bar.Amount,
			Turnover:   bar.Turnover,
			Change:     util.Round2(bar.RisePrice),
			ChangeRate: util.Round2(bar.RiseRate),
			DateTime:   util.FormatKLineDateTime(bar.DateTime, category),
		}
	}
	return out
}

func toStockQuote(item proto.SecurityQuote, symbol string) StockQuote {
	market := types.Market(item.Market).String()
	code := types.CleanCode(item.Code)
	if symbol == "" {
		symbol = fmt.Sprintf("%s.%s", code, market)
	}
	change, changeRate := util.CalcChangeAndRate(item.Price, item.PreClose)
	return StockQuote{
		Code:       code,
		Market:     market,
		Symbol:     symbol,
		Price:      item.Price,
		PreClose:   item.PreClose,
		Change:     change,
		ChangeRate: changeRate,
		Open:       item.Open,
		High:       item.High,
		Low:        item.Low,
		Vol:        item.Vol,
		Amount:     item.Amount,
		Turnover:   item.Turnover,
		ServerTime: item.ServerTime,
		BidLevels:  toQuoteLevels(item.BidLevels),
		AskLevels:  toQuoteLevels(item.AskLevels),
	}
}

func toQuoteLevels(levels []proto.Level) []QuoteLevel {
	if len(levels) == 0 {
		return nil
	}
	out := make([]QuoteLevel, len(levels))
	for i, lv := range levels {
		out[i] = QuoteLevel{Price: lv.Price, Vol: lv.Vol}
	}
	return out
}

func toTopBoardItems(items []proto.TopBoardItem) []TopBoardItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]TopBoardItem, len(items))
	for i, item := range items {
		market := types.Market(item.Market).String()
		code := types.CleanCode(item.Code)
		out[i] = TopBoardItem{
			Code:   code,
			Market: market,
			Symbol: fmt.Sprintf("%s.%s", code, market),
			Price:  item.Price,
			Value:  item.Value,
		}
	}
	return out
}

func selectTopBoard(reply *proto.GetTopBoardReply, boardType int) []proto.TopBoardItem {
	if reply == nil {
		return nil
	}
	switch boardType {
	case TopBoardIncrease:
		return reply.Increase
	case TopBoardDecrease:
		return reply.Decrease
	case TopBoardAmplitude:
		return reply.Amplitude
	case TopBoardRiseSpeed:
		return reply.RiseSpeed
	case TopBoardFallSpeed:
		return reply.FallSpeed
	case TopBoardVolRatio:
		return reply.VolRatio
	case TopBoardPosCommissionRatio:
		return reply.PosCommissionRatio
	case TopBoardNegCommissionRatio:
		return reply.NegCommissionRatio
	case TopBoardTurnover:
		return reply.Turnover
	default:
		return nil
	}
}
