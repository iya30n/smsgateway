package userparam

type IncreaseBalanceRequest struct {
	UserID uint    `json:"user_id"`
	Amount float64 `json:"amount"`
}

type IncreaseBalanceResponse struct {
	NewBalance float64 `json:"new_balance"`
}
