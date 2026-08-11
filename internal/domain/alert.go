package domain

type Alert struct {
	ID        uint64 `json:"id,omitempty" example:"1"`
	Name      string `json:"name" example:"Alex"`
	Message   string `json:"message" example:"Awesome stream!"`
	Type      string `json:"type,omitempty" example:"donation"`
	Duration  int    `json:"duration,omitempty" example:"5000"`
	Timestamp int64  `json:"timestamp" example:"1700000000000"`
}

type AlertRequest struct {
	Name     string `json:"name" example:"Alex" binding:"required"`
	Message  string `json:"message" example:"Awesome stream! Thanks for playing!" binding:"required"`
	Type     string `json:"type,omitempty" example:"donation"`
	Duration int    `json:"duration,omitempty" example:"5000"`
}
