package dto

import "backend-challenge/internal/core/domain"

type BaseResponse[T any] struct {
	Data T `json:"data"`
}

type UserListResponse struct {
	Users []domain.User `json:"users"`
	Total int64         `json:"total"`
}

type SearchLotteryResponse struct {
	Count   int          `json:"count"`
	Items   []LotterItem `json:"items"`
	HasNext bool         `json:"hasNext"`
}

type LotterItem struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Set    int    `json:"set"`
}

func ToSearchLotteryResponse(tickets []domain.LotteryTicket, remaining int64) BaseResponse[SearchLotteryResponse] {
	response := SearchLotteryResponse{
		Count:   len(tickets),
		HasNext: remaining > 0,
	}

	lotteryItems := make([]LotterItem, 0, len(tickets))
	for i := range tickets {
		lotteryItems = append(lotteryItems, LotterItem{
			ID:     tickets[i].ID,
			Number: tickets[i].Number,
			Set:    tickets[i].Set,
		})
	}
	response.Items = lotteryItems

	return BaseResponse[SearchLotteryResponse]{Data: response}

}
