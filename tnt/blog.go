package main

import (
	"fmt"
	"strings"
	"../shared"
)

type BlogFormatter interface {
	FormatPost(post *shared.BlogPost, withComments bool) string
	FormatComment(comment *shared.PostComment) string
}

type TelnetBlogFormatter struct {
	//empty for now
}

func NewTelnetBlogFormatter() *TelnetBlogFormatter {
	return &TelnetBlogFormatter{}
}

func (bf *TelnetBlogFormatter) FormatPost(post *shared.BlogPost, withComments bool) string {
	t := post.Timestamp
	s := fmt.Sprintf("Post #%d, %s\n", post.Id, t.String())
	content := post.Content
	content = strings.Replace(content, "<blockquote>", "\000", -1)
	content = strings.Replace(content, "</blockquote>", "\001", -1)
	content = shared.Htmlstrip(content)
	content = wordwrap(content, 60)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		s += fmt.Sprintf("\t%s\n", line)
	}

	if !withComments {
		return s
	}

	s += fmt.Sprintf("\nComments for post #%d:\n", post.Id)
	for _, c := range post.Comments {
		s += bf.FormatComment(&c)
	}
	return s
}

func (bf *TelnetBlogFormatter) FormatComment(comment *shared.PostComment) string {
	content := shared.Htmlstrip(shared.Telstrip(comment.Content))
	author := shared.Htmlstrip(shared.Telstrip(comment.Author))
	return fmt.Sprintf("\t*[%s] %s\n", author, content)
}
