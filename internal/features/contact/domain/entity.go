package domain

type Address struct {
	Label   string `json:"label"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type Contact struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Emails    []string  `json:"emails"`
	Phones    []string  `json:"phones"`
	Addresses []Address `json:"addresses"`
	Notes     string    `json:"notes"`
}
