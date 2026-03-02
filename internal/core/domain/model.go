package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateUser struct {
	Name     string
	Email    string
	Password string
}

type UpdateUser struct {
	Name  *string
	Email *string
}

type LotteryTicket struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Set    int    `json:"set"`
}
