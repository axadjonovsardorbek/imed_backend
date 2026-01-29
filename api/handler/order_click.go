package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	m "imed/pkg/models"

	"crypto/md5"
	"encoding/hex"
	"math"

	"github.com/gin-gonic/gin"
)

// Click error codes (Supplier side)
const (
	ErrOK              = 0
	ErrSignFailed      = -1
	ErrIncorrectAmount = -2
	ErrActionNotFound  = -3
	ErrAlreadyPaid     = -4
	ErrOrderNotExist   = -5
	ErrTransNotExist   = -6
	ErrFailedUpdate    = -7
	ErrBadRequest      = -8
	ErrCancelled       = -9
)

type clickReq struct {
	ClickTransID      int64
	ServiceID         int
	ClickPaydocID     int64
	MerchantTransID   string
	Amount            float64
	Action            int
	Error             int
	ErrorNote         string
	SignTime          string
	SignString        string
	MerchantPrepareID int64
}

func parseClickForm(c *gin.Context) (clickReq, map[string]string, error) {
	if err := c.Request.ParseForm(); err != nil {
		return clickReq{}, nil, err
	}
	f := func(k string) string { return strings.TrimSpace(c.Request.FormValue(k)) }

	clickTransID, err := strconv.ParseInt(f("click_trans_id"), 10, 64)
	if err != nil {
		return clickReq{}, nil, fmt.Errorf("click_trans_id: %w", err)
	}
	serviceID, err := strconv.Atoi(f("service_id"))
	if err != nil {
		return clickReq{}, nil, fmt.Errorf("service_id: %w", err)
	}
	clickPaydocID, err := strconv.ParseInt(f("click_paydoc_id"), 10, 64)
	if err != nil {
		return clickReq{}, nil, fmt.Errorf("click_paydoc_id: %w", err)
	}
	merchantTransID := f("merchant_trans_id")
	if merchantTransID == "" {
		return clickReq{}, nil, fmt.Errorf("merchant_trans_id missing")
	}
	amount, err := strconv.ParseFloat(f("amount"), 64)
	if err != nil {
		return clickReq{}, nil, fmt.Errorf("amount: %w", err)
	}
	action, err := strconv.Atoi(f("action"))
	if err != nil {
		return clickReq{}, nil, fmt.Errorf("action: %w", err)
	}

	errCode, _ := strconv.Atoi(f("error"))
	errNote := f("error_note")
	signTime := f("sign_time")
	signString := f("sign_string")

	var mpid int64
	if v := f("merchant_prepare_id"); v != "" {
		mpid, _ = strconv.ParseInt(v, 10, 64)
	}

	raw := map[string]string{
		"click_trans_id":      f("click_trans_id"),
		"service_id":          f("service_id"),
		"click_paydoc_id":     f("click_paydoc_id"),
		"merchant_trans_id":   f("merchant_trans_id"),
		"amount":              f("amount"),
		"action":              f("action"),
		"error":               f("error"),
		"error_note":          f("error_note"),
		"sign_time":           f("sign_time"),
		"sign_string":         f("sign_string"),
		"merchant_prepare_id": f("merchant_prepare_id"),
	}

	return clickReq{
		ClickTransID:      clickTransID,
		ServiceID:         serviceID,
		ClickPaydocID:     clickPaydocID,
		MerchantTransID:   merchantTransID,
		Amount:            amount,
		Action:            action,
		Error:             errCode,
		ErrorNote:         errNote,
		SignTime:          signTime,
		SignString:        signString,
		MerchantPrepareID: mpid,
	}, raw, nil
}

// Click response shape
type clickResp struct {
	Error           int    `json:"error"`
	ErrorNote       string `json:"error_note"`
	ClickTransID    int64  `json:"click_trans_id,omitempty"`
	MerchantTransID string `json:"merchant_trans_id,omitempty"`
	MerchantPrepID  int64  `json:"merchant_prepare_id,omitempty"`
	MerchantConfID  int64  `json:"merchant_confirm_id,omitempty"`
}

// isUUID minimal check
func isUUID(s string) bool {
	return len(s) >= 32 && strings.Count(s, "-") >= 4
}

// ---- HANDLERS ----

// POST /click/prepare (action=0)
func (h *Handler) ClickPrepare(c *gin.Context) {
	req, raw, err := parseClickForm(c)
	if err != nil {
		slog.Error("click prepare parse error", "err", err)
		c.JSON(200, clickResp{Error: ErrBadRequest, ErrorNote: "Error in request from click"})
		return
	}

	if req.Action != 0 {
		c.JSON(200, clickResp{Error: ErrActionNotFound, ErrorNote: "Action not found"})
		return
	}

	// service_id check
	if req.ServiceID != h.Cfg.CLICK_SERVICE_ID {
		c.JSON(200, clickResp{Error: ErrBadRequest, ErrorNote: "service_id mismatch"})
		return
	}

	// merchant_trans_id = order_id (UUID)
	if !isUUID(req.MerchantTransID) {
		c.JSON(200, clickResp{Error: ErrOrderNotExist, ErrorNote: "Order not found"})
		return
	}

	// SIGN CHECK
	if !VerifyPrepare(
		h.Cfg.CLICK_SECRET_KEY,
		req.ClickTransID,
		req.ServiceID,
		req.MerchantTransID,
		req.Amount,
		req.Action,
		req.SignTime,
		req.SignString,
	) {
		c.JSON(200, clickResp{Error: ErrSignFailed, ErrorNote: "SIGN CHECK FAILED!"})
		return
	}

	// unique merchant_prepare_id
	merchantPrepareID := time.Now().UnixNano()

	patch := map[string]any{
		"click_prepare_raw": raw,
		"click_prepare_at":  time.Now().Format(time.RFC3339),
	}
	patchJSON, _ := json.Marshal(patch)

	// ✅ context.Background() emas
	res, err := h.Clients.Order.ApplyPrepare(c.Request.Context(), &m.ApplyPrepareParams{
		OrderId:           req.MerchantTransID,
		ClickTransId:      req.ClickTransID,
		ClickPaydocId:     req.ClickPaydocID,
		MerchantPrepareId: merchantPrepareID,
		Amount:            req.Amount,
		MetaPatchJSON:     patchJSON,
	})
	if err != nil {
		slog.Error("click prepare repo error", "err", err)
		c.JSON(200, clickResp{Error: ErrFailedUpdate, ErrorNote: "Failed to update"})
		return
	}

	c.JSON(200, clickResp{
		Error:           res.Error,
		ErrorNote:       res.ErrorNote,
		ClickTransID:    req.ClickTransID,
		MerchantTransID: req.MerchantTransID,
		MerchantPrepID:  res.MerchantPrepareId,
	})
}

// POST /click/complete (action=1)
func (h *Handler) ClickComplete(c *gin.Context) {
	req, raw, err := parseClickForm(c)
	if err != nil {
		slog.Error("click complete parse error", "err", err)
		c.JSON(200, clickResp{Error: ErrBadRequest, ErrorNote: "Error in request from click"})
		return
	}

	if req.Action != 1 {
		c.JSON(200, clickResp{Error: ErrActionNotFound, ErrorNote: "Action not found"})
		return
	}

	if req.ServiceID != h.Cfg.CLICK_SERVICE_ID {
		c.JSON(200, clickResp{Error: ErrBadRequest, ErrorNote: "service_id mismatch"})
		return
	}

	if !isUUID(req.MerchantTransID) {
		c.JSON(200, clickResp{Error: ErrOrderNotExist, ErrorNote: "Order not found"})
		return
	}

	if !VerifyComplete(
		h.Cfg.CLICK_SECRET_KEY,
		req.ClickTransID,
		req.ServiceID,
		req.MerchantTransID,
		req.MerchantPrepareID,
		req.Amount,
		req.Action,
		req.SignTime,
		req.SignString,
	) {
		c.JSON(200, clickResp{Error: ErrSignFailed, ErrorNote: "SIGN CHECK FAILED!"})
		return
	}

	patch := map[string]any{
		"click_complete_raw": raw,
		"click_complete_at":  time.Now().Format(time.RFC3339),
	}
	patchJSON, _ := json.Marshal(patch)

	// ✅ context.Background() emas
	res, err := h.Clients.Order.ApplyComplete(c.Request.Context(), &m.ApplyCompleteParams{
		OrderId:           req.MerchantTransID,
		ClickTransId:      req.ClickTransID,
		MerchantPrepareId: req.MerchantPrepareID,
		ErrorCode:         req.Error,
		MetaPatchJSON:     patchJSON,
	})
	if err != nil {
		slog.Error("click complete repo error", "err", err)
		c.JSON(200, clickResp{Error: ErrFailedUpdate, ErrorNote: "Failed to update"})
		return
	}

	c.JSON(200, clickResp{
		Error:           res.Error,
		ErrorNote:       res.ErrorNote,
		ClickTransID:    req.ClickTransID,
		MerchantTransID: req.MerchantTransID,
		MerchantConfID:  res.MerchantConfirmId,
	})
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Click odatda amountni 2 ta kasr bilan kutadi: "250000.00"
func formatAmount(a float64) string {
	v := math.Round(a*100) / 100
	return fmt.Sprintf("%.2f", v)
}

// action=0
func VerifyPrepare(secret string,
	clickTransID int64,
	serviceID int,
	merchantTransID string,
	amount float64,
	action int,
	signTime string,
	signString string,
) bool {
	raw := fmt.Sprintf("%d%d%s%s%s%d%s",
		clickTransID,
		serviceID,
		secret,
		merchantTransID,
		formatAmount(amount),
		action,
		signTime,
	)
	expected := md5hex(raw)
	return strings.EqualFold(expected, signString)
}

// action=1
func VerifyComplete(secret string,
	clickTransID int64,
	serviceID int,
	merchantTransID string,
	merchantPrepareID int64,
	amount float64,
	action int,
	signTime string,
	signString string,
) bool {
	raw := fmt.Sprintf("%d%d%s%s%d%s%d%s",
		clickTransID,
		serviceID,
		secret,
		merchantTransID,
		merchantPrepareID,
		formatAmount(amount),
		action,
		signTime,
	)
	expected := md5hex(raw)
	return strings.EqualFold(expected, signString)
}
