package domain

type Alert struct {
	ID           uint64            `json:"id,omitempty" example:"1"`
	Name         string            `json:"name,omitempty" example:"Alex"`
	Amount       string            `json:"amount,omitempty" example:"$50.00"`
	Message      string            `json:"message,omitempty" example:"Awesome stream!"`
	Type         string            `json:"type,omitempty" example:"donation"`
	Duration     int               `json:"duration,omitempty" example:"5000"`
	Timestamp    int64             `json:"timestamp" example:"1700000000000"`
	HTMLTemplate string            `json:"html_template,omitempty"`
	CSSStyle     string            `json:"css_style,omitempty"`
	Payload      map[string]string `json:"payload,omitempty"`
}
