package a_stock

import "testing"

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Raw() == nil {
		t.Fatal("Raw() returned nil")
	}
	t.Logf("client ok, address=%q", c.Raw().CurrentAddress())
}

func TestGetAllAStockCodes(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	items, err := c.GetAllAStockCodes("")
	if err != nil {
		t.Fatalf("GetAllAStockCodes() returned error: %v", err)
	}

	t.Logf("共 %d 只 A 股", len(items))
	preview := len(items)
	if preview > 5 {
		preview = 5
	}
	for i := 0; i < preview; i++ {
		t.Logf("[%d] %+v", i, items[i])
	}
}

func TestGetAStockQuotes(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	items, err := c.GetStockQuotesDetail([]string{"300243.SZ"})
	if err != nil {
		t.Fatalf("GetAStockQuotes() returned error: %v", err)
	}
	t.Logf("items: %+v", items)
}

func TestGetStockTopBoard(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	result, err := c.GetStockTopBoard("", 4, 5)
	if err != nil {
		t.Fatalf("GetStockTopBoard() returned error: %v", err)
	}
	t.Logf("%s: %+v", result.Name, result.Items)
}

func TestGetStockUnusual(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	items, err := c.GetStockUnusual("SZ", 0, 2)
	if err != nil {
		t.Fatalf("GetStockUnusual() returned error: %v", err)
	}
	t.Logf("items: %+v", items)
}

func TestGetStockTickChart(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	result, err := c.GetStockTickChart("300532.SZ", 0, 100)
	if err != nil {
		t.Fatalf("GetStockTickChart() returned error: %v", err)
	}
	t.Logf("共 %d 条, 总成交量=%d手, 总成交额=%.2f 元, 均价=%.3f 元", len(result.Items), result.TotalVol, result.TotalAmount, result.AvgPrice/100)
	if len(result.Items) > 0 {
		t.Logf("首条: %+v", result.Items[0])
		t.Logf("尾条: %+v", result.Items[len(result.Items)-1])
	}
}

func TestGetStockHistoryTickChart(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	result, err := c.GetStockHistoryTickChart("300532.SZ", 20260521)
	if err != nil {
		t.Fatalf("GetStockHistoryTickChart() returned error: %v", err)
	}
	t.Logf("共 %d 条, 总成交量=%d手, 总成交额=%.2f 元, 均价=%.3f 元", len(result.Items), result.TotalVol, result.TotalAmount, result.AvgPrice/100)
	if len(result.Items) > 0 {
		t.Logf("首条: %+v", result.Items[0])
		t.Logf("末条: %+v", result.Items[len(result.Items)-1])
	}
}

func TestGetStockKLine(t *testing.T) {
	c := New()
	defer c.Disconnect()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	items, err := c.GetStockKLineSince("300767.SZ", "1d", "2021-01-01")
	if err != nil {
		t.Fatalf("GetStockKLineSince() returned error: %v", err)
	}
	// for i, item := range items {
	// 	t.Logf("[%d] %+v", i, item)
	// }
	t.Logf("共 %d 条", len(items))
	t.Logf("首条: %+v", items[0])
	t.Logf("末条: %+v", items[len(items)-1])
}
