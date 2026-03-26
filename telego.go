package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// concat tc from text and caption of message or post
func tmtc(update tg.Update) (tc string, m *tg.Message) {
	if update.Message == nil {
		return "", nil
	}
	for _, tm := range []*tg.Message{
		update.Message,
		update.EditedMessage,
		update.ChannelPost,
		update.EditedChannelPost,
	} {
		if tm != nil {
			m = tm
			tc += tm.Text + " "
			tc += tm.Caption + " "
			re := tm.ReplyToMessage
			if re != nil {
				tc += re.Text + " "
				tc += re.Caption + " "
			}
			break
		}
	}
	return
}

// match pattern in text and caption of message or post
func anyWithMatch(pattern *regexp.Regexp) th.Predicate {
	return func(ctx context.Context, update tg.Update) bool {
		tc, _ := tmtc(update)
		return pattern.MatchString(tc)
	}
}

// is command in text and caption of update
func AnyCommand() th.Predicate {
	return func(ctx context.Context, update tg.Update) bool {
		_, ctm := tmtc(update)
		if ctm == nil {
			return false
		}
		return strings.HasPrefix(ctm.Text, "/") || strings.HasPrefix(ctm.Caption, "/")
	}
}

// is command in text and caption of update
func leftChat() th.Predicate {
	return func(ctx context.Context, update tg.Update) bool {
		return update.Message != nil &&
			update.Message.LeftChatMember != nil
	}
}

// is message about new member
func newMember() th.Predicate {
	return func(ctx context.Context, update tg.Update) bool {
		return update.Message != nil &&
			len(update.Message.NewChatMembers) > 0
	}
}

// is reply to bot message as delete command
func ReplyMessageIsMinus() th.Predicate {
	return func(ctx context.Context, update tg.Update) bool {
		return update.Message != nil &&
			update.Message.ReplyToMessage != nil &&
			update.Message.Text == "-"
	}
}

// telego logger interface
type Logger struct{}

// hide bot token
func woToken(format string, args ...any) (s string) {
	s = src(10) + " " + fmt.Sprintf(format, args...)
	btStart := strings.Index(s, "/bot") + 4
	if btStart > 4-1 {
		btLen := strings.Index(s[btStart:], "/")
		if btLen > 0 {
			s = s[:btStart] + s[btStart+btLen:]
		}
	}
	return
}

// bot debug message
func (Logger) Debugf(format string, args ...any) {
	lt.Print(woToken(format, args...))
}

// bot error message
func (Logger) Errorf(format string, args ...any) {
	let.Print(woToken(format, args...))
}

// send error message to first chatID in args
func SendError(bot *tg.Bot, err error) {
	if bot != nil && len(chats) > 0 && err != nil {
		bot.SendMessage(mainCtx, tu.MessageWithEntities(tu.ID(chats[0]),
			tu.Entity("💥"),
			tu.Entity(err.Error()).Code(),
		))
	}
}
