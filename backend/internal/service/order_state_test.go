package service

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from string
		to   string
		want bool
	}{
		{"pending_payment", "pending_shipment", true}, // 付款
		{"pending_shipment", "shipped", true},         // 发货
		{"shipped", "received", true},                 // 收货
		{"received", "completed", true},               // 完成
		{"pending_payment", "cancelled", true},        // 取消
		{"pending_shipment", "cancelled", true},       // 卖家取消
		{"shipped", "cancelled", false},               // 已发货不可取消
		{"completed", "shipped", false},               // 非法回退
		{"pending_payment", "received", false},        // 跳步
	}
	for _, tc := range cases {
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			if got := canTransition(tc.from, tc.to); got != tc.want {
				t.Fatalf("canTransition(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestGenOrderNo(t *testing.T) {
	a := genOrderNo()
	b := genOrderNo()
	if a == "" || b == "" || a == b {
		t.Fatalf("order no should be non-empty and unique, got %q %q", a, b)
	}
}
