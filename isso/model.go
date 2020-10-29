package isso

// Thread is comments thread
type Thread struct {
	ID    int64
	URI   string `validate:"required,uri"`
	Title string
}

// Comment is comment saved in database
type Comment struct {
	ID           int64     `json:"id"`
	Parent       *int64    `json:"parent"`
	Created      float64   `json:"created"`
	Modified     *float64  `json:"modified"`
	Mode         int       `json:"mode"`
	Text         string    `json:"text"  validate:"required,gte=3,lte=65535"`
	Author       string    `json:"author"`
	Email        *string   `json:"email,omitempty"  validate:"omitempty,email"`
	Website      *string   `json:"website"  validate:"omitempty,url"`
	Likes        int       `json:"likes"`
	Dislikes     int       `json:"dislikes"`
	Notification int       `json:"notification" validate:"omitempty,min=0,max=2"`
	RemoteAddr   string    `json:"-" validate:"required,ip"`
	Voters       [256]byte `json:"-"`
}

type submittedComment struct {
	Comment
	URI   string `json:"-" validate:"required,uri"`
	Title string `json:"title" validate:"omitempty"`
	Key   string `json:"key"`
	Sig   string `json:"sig"`
	K1    string `json:"k1"`
}

type reply struct {
	Comment
	Hash          string   `json:"hash"`
	HiddenReplies *int64   `json:"hidden_replies,omitempty"`
	TotalReplies  *int64   `json:"total_replies,omitempty"`
	Replies       *[]reply `json:"replies,omitempty"`
}

// Convert remove email from comment, and markdownify if not `plain`
// if markdown convert failed, c.Text will be origin text. but return error is not nil
func (c Comment) convert(plain bool, hash interface{ Hash(string) string }, markdown interface {
	Convert(source string) (string, error)
}) (reply, error) {

	// hash comment
	var hashresult string
	if c.Email != nil {
		hashresult = hash.Hash(*c.Email)
	} else {
		hashresult = hash.Hash(c.RemoteAddr)
	}

	// remove email
	c.Email = nil

	// markdowify
	if plain {
		return reply{c, hashresult, nil, nil, nil}, nil
	}
	text, err := markdown.Convert(c.Text)
	if err != nil {
		return reply{c, hashresult, nil, nil, nil}, err
	}
	c.Text = text
	return reply{c, hashresult, nil, nil, nil}, nil
}
