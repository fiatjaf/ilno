package ilno

// Thread is comments thread
type Thread struct {
	ID    int64
	URI   string `validate:"required,uri"`
	Title string
}

// Comment is comment saved in database
type Comment struct {
	ID       int64     `json:"id"`
	Parent   *int64    `json:"parent"`
	Created  float64   `json:"created"`
	Modified *float64  `json:"modified"`
	Mode     int       `json:"mode"`
	Text     string    `json:"text" validate:"required,gte=3,lte=65535"`
	Key      string    `json:"key" validate:"required"`
	Author   string    `json:"author"`
	Likes    int       `json:"likes"`
	Dislikes int       `json:"dislikes"`
	Voters   [256]byte `json:"-"`
}

type submittedComment struct {
	Comment
	URI   string `json:"-" validate:"required,uri"`
	Title string `json:"title" validate:"omitempty"`
	Sig   string `json:"sig"`
	K1    string `json:"k1"`
}

type reply struct {
	Comment
	HiddenReplies *int64   `json:"hidden_replies,omitempty"`
	TotalReplies  *int64   `json:"total_replies,omitempty"`
	Replies       *[]reply `json:"replies,omitempty"`
}
