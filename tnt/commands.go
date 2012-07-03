package main

import (
	//	"net"
	//	"os"
	"fmt"
	//	"bufio"
	"strconv"
	"strings"
	"time"
	"../shared"
	//	"crypto/md5"
)

type BlogCommand struct {
	handler        func(*BlogSession, []string) string
	min_perm_level int //the permission level needed to execute this command
}

type BlogCommandHandler interface {
	AddCommand(state int, commandString string, command BlogCommand)
	HandleCommand(session *BlogSession, commandline []string) string
}

type CommandMap map[string]BlogCommand

type TelnetCommandHandler struct {
	commandsByState map[int]CommandMap
}

func NewTelnetCommandHandler() *TelnetCommandHandler {
	h := &TelnetCommandHandler{}
	h.commandsByState = make(map[int]CommandMap)

	h.commandsByState[state_reading] = make(CommandMap)
	h.commandsByState[state_posting] = make(CommandMap)

	h.setupCMDHandlers()
	return h
}

func (h *TelnetCommandHandler) AddCommand(state int, commandString string, command BlogCommand) {
	cm := h.commandsByState[state]
	cm[commandString] = command
}

func (h *TelnetCommandHandler) HandleCommand(session *BlogSession, commandline []string) string {
	state := session.State()
	cmdmap := h.commandsByState[state]

	//handle normal reading mode
	k, ok := cmdmap[commandline[0]]
	if !ok {
		//if user is posting we don't want to send error messages for his input
		if session.State() != state_posting {
			return "error: command not implemented\n"
		} else {
			return ""
		}
	}
	if session.PermissionLevel() >= k.min_perm_level {
		handler := k.handler
		return handler(session, commandline)
		//		session.Send( handler(session, items) )
	} else {
		//		session.Send("error: privileges too low\n")
		return "error: privileges too low\n"
	}
	return "\n"
}

func (h *TelnetCommandHandler) setupCMDHandlers() {

	f := func(session *BlogSession, items []string) string {
		session.Disconnect()
		return "ok\n"
	}
	h.AddCommand(state_reading, "quit", BlogCommand{f, 0})

	f = func(session *BlogSession, items []string) string {
		session.Disconnect()
		session.Server().Shutdown()
		return "ok\n"
	}
	h.AddCommand(state_reading, "die", BlogCommand{f, 12})

	h.AddCommand(state_reading, "auth",
		BlogCommand{
			handler:        tch_handleAuth,
			min_perm_level: 0,
		})

	h.AddCommand(state_reading, "", BlogCommand{tch_handleNullspace, 0})
	h.AddCommand(state_reading, "read", BlogCommand{tch_handleRead, 0})
	h.AddCommand(state_reading, "news", BlogCommand{tch_handleNews, 0})
	h.AddCommand(state_reading, "today", BlogCommand{tch_handleToday, 0})
	h.AddCommand(state_reading, "post", BlogCommand{tch_handlePost, 12})
	h.AddCommand(state_reading, "comment", BlogCommand{tch_handleComment, 0})
	h.AddCommand(state_reading, "broadcast", BlogCommand{tch_handleBroadcast, 0})
	h.AddCommand(state_reading, "help", BlogCommand{tch_handleHelp, 0})
	h.AddCommand(state_posting, "$end", BlogCommand{tch_handlePostingEnd, 0})

}

func tch_handleRead(session *BlogSession, items []string) string {
	if len(items) != 2 {
		return "syntax: read <post_id>\n"
	}
	db := shared.DBGet()
	defer db.Close()

	id, _ := strconv.ParseInt(items[1], 10, 64)
	post, err := db.GetPost(id)
	if err != nil {
		return err.Error() + "\n"
	}

	post.Comments, _ = db.GetComments(post.Id)

	return session.BlogFormatter().FormatPost(&post, true)
}

func tch_handleAuth(session *BlogSession, items []string) string {
	if len(items) != 2 {
		return "syntax: auth <password>\n"
	}
	password := items[1]
	b := session.Auth(password)

	if !b {
		return "couldn't change permission level\n"
	}
	return fmt.Sprintf("permission level %d granted\n", session.PermissionLevel())
}

func tch_handlePost(session *BlogSession, items []string) string {
	if len(items) != 1 {
		return "syntax: post\n"
	}
	session.ResetInputBuffer()
	session.SetState(state_posting)
	return "enter post. enter $end to end input and save post.\n01234567890123456789012345678901234567890123456789012345678901234567890123456789\n"
}

func tch_handleComment(session *BlogSession, items []string) string {
	if len(items) < 4 {
		return "syntax: comment <post_id> <your_nick> <your many words of comment>\n"
	}

	post_id, _ := strconv.ParseInt(items[1], 10, 64)
	nick := items[2]
	content := strings.Join(items[3:], " ")

	comment := shared.PostComment{
		Content:   content,
		Author:    nick,
		Timestamp: time.Now(),
		PostId:    post_id,
	}

	db := shared.DBGet()
	defer db.Close()

	i, err := db.StoreComment(&comment)
	if err != nil {
		return "error: " + err.Error() + "\n"
	}

	s := fmt.Sprintf("commented post %d. your comment's id: %d\n", post_id, i)
	return s
}

func tch_handleBroadcast(session *BlogSession, items []string) string {
	if len(items) < 2 {
		return "syntax: broadcast <your broadcast>\n"
	}

	message := strings.Join(items[1:], " ")
	message += "\n"
	session.Server().Broadcast(message)

	return "Broadcast sent\n"
}

func tch_handlePostingEnd(session *BlogSession, items []string) string {
	session.SetState(state_reading)

	post := shared.BlogPost{
		Content:   strings.Trim(strings.Replace(session.InputBuffer(), "$end", "", -1), "\n\r"),
		Timestamp: time.Now(),
		Id:        0, //0 = create new post
	}

	db := shared.DBGet()
	defer db.Close()

	id, err := db.StorePost(&post)
	if err != nil {
		return "error: " + err.Error() + "\n"
	}

	s := fmt.Sprintf("saved post with id %d\n", id)
	return s
}

func tch_handleNews(session *BlogSession, items []string) string {
	if len(items) < 1 || len(items) > 2 {
		return "syntax: news [max_number_of_posts]\n"
	}

	num := 5
	if len(items) == 2 {
		num, _ = strconv.Atoi(items[1])
	}

	db := shared.DBGet()
	defer db.Close()

	posts, err := db.GetLastNPosts(int32(num))
	if err != nil {
		return err.Error() + "\n"
	}

	//post.Comments, _ = db.GetComments(post.Id)
	s := ""
	for _, post := range posts {
		//s += session.BlogFormatter().FormatPost(&post, false)
		//s += "\n"
		s = session.BlogFormatter().FormatPost(&post, false) + s
		s = "\n" + s
	}

	return s
}

func tch_handleToday(session *BlogSession, items []string) string {
	if len(items) != 1 {
		return "syntax: today\n"
	}

	today_t := time.Now()
	today_t = time.Date(today_t.Year(), today_t.Month(), today_t.Day(), 0, 0, 0, 0, today_t.Location())

	today := today_t.Unix()
	tomorrow := today + (24 * 60 * 60)

	//GetPostsForTimespan(start_timestamp, end_timestamp int64) (posts []BlogPost, err os.Error)

	//    fmt.Printf("today: %d | tomorro: %d\n", today, tomorrow)

	db := shared.DBGet()
	defer db.Close()

	posts, err := db.GetPostsForTimespan(today, tomorrow, 1)
	if err != nil {
		return err.Error() + "\n"
	}

	//post.Comments, _ = db.GetComments(post.Id)
	s := ""
	for _, post := range posts {
		s += session.BlogFormatter().FormatPost(&post, false)
		s += "\n"
	}

	return s
}

func tch_handleNullspace(session *BlogSession, items []string) string {
	return ""
}

func tch_handleHelp(session *BlogSession, items []string) string {

	s := "fettemama help\n"
	s += "help\n\t* this screen\n"
	s += "comment <post_id> <your_nick> ...\n\t* add comment to a post\n"
	s += "post\n\t* create new blog post\n"
	s += "auth <password>\n\t* change user level\n"
	s += "read <post_id>\n\t* read a post\n"
	s += "news [num of posts]\n\t* shows the last num posts\n"
	s += "today\n\t* shows today's posts\n"
	s += "broadcast <your broadcast messega>\n\t* sends a message to all logged in users\n"
	s += "quit\n\t* ends your session\n"

	s += "\n"
	return s
}
