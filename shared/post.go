package shared
import (
  "time"
)

type BlogPost struct {
	Content   string
	//Timestamp int64
	Timestamp time.Time
	Id        int64
	Comments  []PostComment
}

func (self BlogPost) Excerpt() string {
	t := Htmlstrip(self.Content)
	l := 80
	if len(t) <= 80 {
		l = len(t)
	}
	ex := t[0:l]
	return string(ex)
}

type PostComment struct {
	Content   string
	Author    string
	//Timestamp int64
	Timestamp time.Time
	Id        int64
	PostId    int64
}
