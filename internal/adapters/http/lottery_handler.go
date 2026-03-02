package http

import (
	"log/slog"
	"net/http"

	"backend-challenge/internal/adapters/http/dto"
	"backend-challenge/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type LotteryHandler struct {
	lottery ports.LotteryService
}

func NewLotteryHandler(lottery ports.LotteryService) *LotteryHandler {
	return &LotteryHandler{lottery: lottery}
}

// SearchLottery searches for lottery tickets by pattern
//
//	@Summary		Search lottery tickets
//	@Description	Search for available lottery tickets using a 6-digit pattern (e.g., '123***', '******'). Returns a list of matching tickets and a flag indicating if more results are available.
//	@Tags			lottery
//	@Produce		json
//	@Param			pattern	query		string	true	"6-digit pattern (e.g., 123***)"
//	@Param			limit	query		int		false	"Number of tickets to return (default 10, max 100)"
//	@Success		200		{object}	dto.SearchLotteryResponse
//	@Failure		400		{object}	map[string]string	"Invalid pattern"
//	@Failure		404		{object}	map[string]string	"No tickets available"
//	@Router			/lottery/search [get]
func (h *LotteryHandler) SearchLottery(c *gin.Context) {
	var query dto.LotterySearchRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		writeValidationError(c, err)
		return
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 1
	}

	tickets, remaining, err := h.lottery.Search(c.Request.Context(), query.Pattern, limit)
	if err != nil {
		slog.Error("cannot search lottery",
			"error", err,
		)
		handleError(c, err)
		return
	}

	writeJSON(c, http.StatusOK, dto.ToSearchLotteryResponse(tickets, remaining))
}
