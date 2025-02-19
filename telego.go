package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fasthttp/router"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
	"golang.ngrok.com/ngrok"
	nc "golang.ngrok.com/ngrok/config"
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
	return func(update tg.Update) bool {
		tc, _ := tmtc(update)
		return pattern.MatchString(tc)
	}
}

// is command in text and caption of update
func AnyCommand() th.Predicate {
	return func(update tg.Update) bool {
		_, ctm := tmtc(update)
		if ctm == nil {
			return false
		}
		return strings.HasPrefix(ctm.Text, "/") || strings.HasPrefix(ctm.Caption, "/")
	}
}

// is command in text and caption of update
func leftChat() th.Predicate {
	return func(update tg.Update) bool {
		return update.Message != nil &&
			update.Message.LeftChatMember != nil
	}
}

// is message about new member
func newMember() th.Predicate {
	return func(update tg.Update) bool {
		return update.Message != nil &&
			len(update.Message.NewChatMembers) > 0
	}
}

// is reply to bot message as delete command
func ReplyMessageIsMinus() th.Predicate {
	return func(update tg.Update) bool {
		return update.Message != nil &&
			update.Message.ReplyToMessage != nil &&
			update.Message.Text == "-"
	}
}

// func Delete(ChatID tg.ChatID, MessageID int) *tg.DeleteMessageParams {
// 	return &tg.DeleteMessageParams{
// 		ChatID:    ChatID,
// 		MessageID: MessageID,
// 	}
// }

// set secretToken to FastHTTPWebhookServer and SetWebhookParams
func UpdatesWithSecret(b *tg.Bot, secretToken, publicURL, endPoint string) (<-chan tg.Update, error) {
	whs := tg.FastHTTPWebhookServer{
		Logger:      b.Logger(),
		Server:      &fasthttp.Server{},
		Router:      router.New(),
		SecretToken: secretToken,
	}
	whp := &tg.SetWebhookParams{
		URL:         publicURL + endPoint,
		SecretToken: secretToken,
	}
	return b.UpdatesViaWebhook(endPoint,
		tg.WithWebhookServer(whs),
		tg.WithWebhookSet(whp))
}

// start ngrok.Tunnel with NGROK_AUTHTOKEN in env (optional) and SecretToken
func UpdatesWithNgrok(b *tg.Bot, secretToken, endPoint string) (<-chan tg.Update, error) {
	var (
		err error
		tun ngrok.Tunnel
	)
	// If NGROK_AUTHTOKEN in env and account is free and is already open need return
	// else case ngrok.Listen hang
	ctx, ca := context.WithTimeout(context.Background(), time.Second)
	sess, err := ngrok.Connect(ctx, ngrok.WithAuthtokenFromEnv()) //even without NGROK_AUTHTOKEN in env
	if err != nil {
		return nil, Errorf("tunnel already open %w", err)
	}
	sess.Close()
	ca()

	ctx, ca = context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			ca()
		}
	}()
	tun, err = ngrok.Listen(
		ctx,
		nc.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return nil, srcError(err)
	}
	publicURL := tun.URL()
	if secretToken == "" {
		secretToken = tun.ID()
	}
	if endPoint == "" {
		endPoint = "/" + secretToken
	}

	whs := tg.FastHTTPWebhookServer{
		Logger:      b.Logger(),
		Server:      &fasthttp.Server{},
		Router:      router.New(),
		SecretToken: secretToken,
	}
	whp := &tg.SetWebhookParams{
		URL:         publicURL + endPoint,
		SecretToken: secretToken,
	}
	fws := tg.FuncWebhookServer{
		Server: whs,
		// Override default start func to use Ngrok tunnel
		StartFunc: func(address string) error {
			ltf.Println("StartFunc", address)
			err := whs.Server.Serve(tun) //always return error
			if err.Error() == "failed to accept connection: Tunnel closed" {
				ltf.Println("Serve ok")
				return nil
			}
			letf.Println("Serve", err)
			return srcError(err)
		},
		// Override default stop func to close Ngrok tunnel
		StopFunc: func(_ context.Context) error {
			ltf.Println("StopFunc")
			ca() //need for NGROK_AUTHTOKEN in env
			return nil
		},
	}
	return b.UpdatesViaWebhook(endPoint,
		tg.WithWebhookServer(fws),
		tg.WithWebhookSet(whp))
}

// start ngrok.Tunnel with NGROK_AUTHTOKEN in env (optional) and SecretToken not used serve but loop of accept
func UpdatesWithNgrokAccept(b *tg.Bot, secretToken, endPoint string) (<-chan tg.Update, error) {
	var (
		err error
		tun ngrok.Tunnel
	)
	// If NGROK_AUTHTOKEN in env and account is free and is already open need return
	// else case ngrok.Listen hang
	ctx, ca := context.WithTimeout(context.Background(), time.Second)
	sess, err := ngrok.Connect(ctx, ngrok.WithAuthtokenFromEnv()) //even without NGROK_AUTHTOKEN in env
	if err != nil {
		return nil, Errorf("tunnel already open %w", err)
	}
	sess.Close()
	ca()

	ctx, ca = context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			ca()
		}
	}()
	tun, err = ngrok.Listen(
		ctx,
		nc.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return nil, srcError(err)
	}
	publicURL := tun.URL()
	if secretToken == "" {
		secretToken = tun.ID()
	}
	if endPoint == "" {
		endPoint = "/" + secretToken
	}
	b.Logger().Debugf("%s %s %s %s", publicURL, tun.ForwardsTo(), secretToken, endPoint)

	whs := tg.FastHTTPWebhookServer{
		Logger:      b.Logger(),
		Server:      &fasthttp.Server{},
		Router:      router.New(),
		SecretToken: secretToken,
	}
	whp := &tg.SetWebhookParams{
		URL:         publicURL + endPoint,
		SecretToken: secretToken,
	}
	fws := tg.FuncWebhookServer{
		Server: whs,
		// Override default stop func to close Ngrok tunnel
		StopFunc: func(_ context.Context) error {
			b.Logger().Debugf("StopFunc")
			ca() //need for NGROK_AUTHTOKEN in env
			return nil
		},
	}

	go func() {
		for {
			conn, err := tun.Accept()
			if err != nil {
				b.Logger().Errorf("tun.Accept %v", err)
				return
			}
			b.Logger().Debugf("%s => %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
			go func() {
				err := whs.Server.ServeConn(conn)
				if err != nil {
					b.Logger().Errorf("Server.ServeConn(%v): %v", conn, err)
				}
				b.Logger().Debugf("Server.ServeConn ok")
			}()
		}
	}()

	return b.UpdatesViaWebhook(endPoint,
		tg.WithWebhookServer(fws),
		tg.WithWebhookSet(whp))
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
		bot.SendMessage(tu.MessageWithEntities(tu.ID(chats[0]),
			tu.Entity("💥"),
			tu.Entity(err.Error()).Code(),
		))
	}
}
