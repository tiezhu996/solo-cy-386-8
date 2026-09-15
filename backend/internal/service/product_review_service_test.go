package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/util"
)

func newTestReviewServices() (*ProductService, *ProductReviewService, *fakeProductRepo, *fakeProductReviewRepo) {
	pr := newFakeProductRepo()
	fr := newFakeFavoriteRepo()
	rr := newFakeProductReviewRepo()
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	ps := NewProductService(nil, pr, fr, rr, logger)
	rs := NewProductReviewService(nil, rr, pr, logger)
	return ps, rs, pr, rr
}

func appErrorCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *util.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T %v", err, err)
	}
	return ae.Code
}

func TestReviewDecideLifecycle(t *testing.T) {
	ps, rs, pr, rr := newTestReviewServices()
	product, err := ps.Create(1, dto.ProductCreateRequest{
		Title: "Switch", Description: "二手游戏机，配件齐全", OriginalPrice: 2099, Price: 1299,
		Condition: "almost_new", Category: "digital",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	rv, err := rr.GetPendingByProductForUpdate(nil, product.ID)
	if err != nil {
		t.Fatalf("pending review: %v", err)
	}
	decide := func(reviewerID, id uint, req dto.ProductReviewDecisionRequest) error {
		_, err := rs.Decide(reviewerID, id, req)
		return err
	}

	// 驳回不填原因 → CodeReviewReasonEmpty。
	if code := appErrorCode(t, decide(99, rv.ID, dto.ProductReviewDecisionRequest{Action: "reject"})); code != constants.CodeReviewReasonEmpty {
		t.Fatalf("expected reason empty code %d, got %d", constants.CodeReviewReasonEmpty, code)
	}

	// 正式驳回。
	if err := decide(99, rv.ID, dto.ProductReviewDecisionRequest{Action: "reject", Reason: "描述与图片不符"}); err != nil {
		t.Fatalf("reject: %v", err)
	}
	p := pr.products[product.ID]
	if p.ReviewStatus != constants.ProductReviewRejected || p.Status != constants.ProductStatusOffShelf {
		t.Fatalf("expected rejected/off_shelf, got %s/%s", p.ReviewStatus, p.Status)
	}
	if p.RejectReason != "描述与图片不符" {
		t.Fatalf("expected reject reason persisted, got %q", p.RejectReason)
	}

	// 重复审核同一轮记录 → 状态冲突，不能重复改状态。
	if code := appErrorCode(t, decide(100, rv.ID, dto.ProductReviewDecisionRequest{Action: "approve"})); code != constants.CodeReviewStateInvalid {
		t.Fatalf("expected state invalid code %d, got %d", constants.CodeReviewStateInvalid, code)
	}

	// 驳回后卖家修改重提。
	desc := "已更换实拍图"
	if _, err := ps.Update(1, product.ID, dto.ProductUpdateRequest{Description: &desc}); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	rv2, err := rr.GetPendingByProductForUpdate(nil, product.ID)
	if err != nil || rv2.Round != 2 {
		t.Fatalf("expected round 2 pending review, got %+v err=%v", rv2, err)
	}

	// 复审通过 → 上架可售。
	if err := decide(99, rv2.ID, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	p = pr.products[product.ID]
	if p.ReviewStatus != constants.ProductReviewApproved || p.Status != constants.ProductStatusOnSale {
		t.Fatalf("expected approved/on_sale, got %s/%s", p.ReviewStatus, p.Status)
	}

	// 已通过的记录再次审核同样被拒（重复审核不能重复改状态）。
	if code := appErrorCode(t, decide(100, rv2.ID, dto.ProductReviewDecisionRequest{Action: "reject", Reason: "x"})); code != constants.CodeReviewStateInvalid {
		t.Fatalf("duplicate decide should conflict, got code %d", code)
	}
}

func TestReviewStateMachineTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"", constants.ProductReviewPending, true},
		{constants.ProductReviewApproved, constants.ProductReviewPending, true}, // 修改已上架 → 复审
		{constants.ProductReviewRejected, constants.ProductReviewPending, true}, // 驳回后重提
		{constants.ProductReviewPending, constants.ProductReviewApproved, true}, // 通过
		{constants.ProductReviewPending, constants.ProductReviewRejected, true}, // 驳回
		{constants.ProductReviewPending, constants.ProductReviewPending, false}, // 重复提交
		{constants.ProductReviewApproved, constants.ProductReviewRejected, false},
	}
	for _, tc := range cases {
		if got := constants.CanReviewTransition(tc.from, tc.to); got != tc.want {
			t.Fatalf("CanReviewTransition(%s,%s)=%v want %v", tc.from, tc.to, got, tc.want)
		}
	}
}
