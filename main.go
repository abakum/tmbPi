package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/jibber_jabber"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/xlab/closer"
)

func main() {
	mainCtx, mainCancel = context.WithCancel(context.Background())
	ttCtx, ttCancel = context.WithCancel(mainCtx)

	var (
		err error
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			SendError(bot, err)
			defer os.Exit(1)
		}
		ltf.Println("closer mainCancel()")
		mainCancel()
		ltf.Println("closer ips.close")
		ips.close()
		wg.Wait()
		// pressEnter()
	})
	ul, err = jibber_jabber.DetectLanguage()
	if err != nil {
		ul = "en"
	}
	if len(chats) == 0 {
		err = Errorf(dic.add(ul,
			"en:Usage: %s AllowedChatID1 AllowedChatID2 AllowedChatIDx\n",
			"ru:Использование: %s РазрешённыйChatID1 РазрешённыйChatID2 РазрешённыйChatIDх\n",
		), os.Args[0])
		return
	} else {
		li.Println(dic.add(ul,
			"en:Allowed ChatID:",
			"ru:Разрешённые ChatID:",
		), chats)
	}
	ex, err := os.Getwd()
	if err == nil {
		tmbPingJson = filepath.Join(ex, tmbPingJson)
	}
	li.Println(filepath.FromSlash(tmbPingJson))

	bot, err = CreateBotWithProxy(os.Getenv("TOKEN"))

	if err != nil {
		if errors.Is(err, tg.ErrInvalidToken) {
			err = Errorf(dic.add(ul,
				"en:set TOKEN=BOT_TOKEN",
				"ru:Присвойте BOT_TOKEN переменной окружения TOKEN",
			))
		}
		err = srcError(err)
		return
	}

	me, err = bot.GetMe(mainCtx)
	if err != nil {
		err = srcError(err)
		return
	}

	// bot.DeleteMyCommands(nil)
	tacker = time.NewTicker(tt)
	defer tacker.Stop()
	bh, err := startH(ttCtx)
	SendError(bot, fmt.Errorf("startH %v", err))
	if err != nil {
		return
	}

	wg.Add(1)
	go saver()

	wg.Add(1)
	// main loop
	go func() {
		defer wg.Done()
		ticker = time.NewTicker(dd)
		defer ticker.Stop()
		// tacker = time.NewTicker(tt)
		defer tacker.Stop()
		for {
			select {
			case <-mainCtx.Done():
				ltf.Println("Ticker done")
				return
			case t := <-ticker.C:
				ltf.Println("Tick at", t)
				ips.update(customer{})
			case t := <-tacker.C:
				ltf.Println("Tack at", t)
				SendError(bot, fmt.Errorf("stopH %v", stopH(ttCancel, bh)))
				ttCtx, ttCancel = context.WithCancel(mainCtx)
				bh, err = startH(ttCtx)
				SendError(bot, fmt.Errorf("startH %v", err))
				if err != nil {
					letf.Println(err)
					restart(tacker, tt)
				}
			}
		}
	}()

	err = loader()
	if err != nil {
		return
	}
	closer.Hold()
}

// stop handler, webhook, polling
func stopH(cancel context.CancelFunc, bh *th.BotHandler) (err error) {

	if cancel != nil {
		ltf.Println("StopLongPolling")
		cancel()
	}
	if bh != nil {
		ltf.Println("bh.Stop")
		err = srcError(bh.Stop())
	}
	return
}

// start handler and webhook or polling
func startH(ctx context.Context) (*th.BotHandler, error) {
	// updates, err = bot.UpdatesViaLongPolling(nil)
	updates, err := bot.UpdatesViaLongPolling(ctx, &tg.GetUpdatesParams{Timeout: int(refresh.Seconds())})
	if err != nil {
		return nil, srcError(err)
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return nil, srcError(err)
	}

	//AnyCallbackQueryWithMessage
	bh.Handle(bhAnyCallbackQueryWithMessage, th.AnyCallbackQueryWithMessage())
	//delete reply message with - or / in text
	bh.Handle(bhReplyMessageIsMinus, ReplyMessageIsMinus())
	//anyWithIP
	bh.Handle(bhAnyWithMatch, anyWithMatch(reIP))
	//AnyCommand
	bh.Handle(bhAnyCommand, AnyCommand())
	//leftChat
	bh.Handle(bhLeftChat, leftChat())
	//newMember
	bh.Handle(bhNewMember, newMember())
	//anyWithYYYYMMDD Easter Egg expected "name YYYY.?MM.?DD"
	bh.Handle(bhEasterEgg, anyWithMatch(reYYYYMMDD))

	go bh.Start()

	return bh, nil
}

// handler IP
func bhAnyWithMatch(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	tc, ctm := tmtc(update)
	if ctm == nil {
		return nil
	}
	ok, ups := allowed(ul, ctm.From.ID, ctm.Chat.ID)
	keys, _ := set(reIP.FindAllString(tc, -1))
	ltf.Println("bh.Handle anyWithIP", keys, ctm)
	if ok {
		for _, ip := range keys {
			ips.write(ip, customer{Tm: ctm})
		}
	} else {
		ikbsf = len(ikbs) - 1
		news := ""
		for _, ip := range keys {
			if ips.read(ip) {
				ips.write(ip, customer{Tm: ctm})
			} else {
				news += ip + " "
			}
		}
		if len(news) > 1 {
			params := tu.MessageWithEntities(tu.ID(ctm.Chat.ID),
				tu.Entity("/"+strings.TrimRight(news, " ")).Code(),
				tu.Entity(ups),
			)
			params.ReplyParameters = &tg.ReplyParameters{MessageID: ctm.MessageID}
			params.ReplyMarkup = tu.InlineKeyboard(tu.InlineKeyboardRow(ikbs[ikbsf:]...))
			_, err := bot.SendMessage(mainCtx, params)
			if err != nil {
				let.Println(err)
			}

		}
		return nil
	}
	return nil
}

// handler EasterEgg
func bhEasterEgg(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	tc, ctm := tmtc(update)
	if ctm == nil {
		return nil
	}
	if ctm.Chat.Type != "private" {
		return nil
	}
	keys, _ := set(reYYYYMMDD.FindAllString(tc, -1))
	ltf.Println("bh.Handle anyWithYYYYMMDD", keys)
	for _, key := range keys {
		fss := reYYYYMMDD.FindStringSubmatch(key)
		bd, err := time.ParseInLocation("20060102150405", strings.Join(fss[2:], "")+"120000", time.Local)
		if err == nil {
			nbd := fmt.Sprintf("%s %s", fss[1], bd.Format("2006-01-02"))
			tl := start(me, nbd)
			entitys := []tu.MessageEntityCollection{tu.Entity("⚡").TextLink(tl)}
			entitys = append(entitys, tu.Entity(nbd).Code())
			entitys = append(entitys, tu.Entity("🔗"+"\n").TextLink("t.me/share/url?url="+tl))
			le := len(entitys) + 1
			for _, year := range la(bd) {
				// b, a, ok := strings.Cut(year, " ")
				// entitys = append(entitys, tu.Entity(b).Hashtag())
				// if ok {
				// 	entitys = append(entitys, tu.Entityf(" n%s\n", a))
				// }
				entitys = append(entitys, tu.Entity(year+"\n"))
			}
			if len(entitys) > le {
				entitys[len(entitys)-1] = entitys[len(entitys)-1].Spoiler()
			}
			params := tu.MessageWithEntities(tu.ID(ctm.Chat.ID), entitys...)
			params.ReplyParameters = &tg.ReplyParameters{MessageID: ctm.MessageID}
			_, err := bot.SendMessage(mainCtx, params)
			if err != nil {
				let.Println(err)
			}

		}
	}
	return nil
}

// handler Callback
func bhAnyCallbackQueryWithMessage(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	uc := update.CallbackQuery
	if uc == nil {
		return nil
	}
	tm := uc.Message
	if tm == nil {
		return nil
	}
	if !tm.IsAccessible() {
		return nil
	}
	message := tm.(*tg.Message)
	my := true
	if message.Chat.Type != "private" && message.ReplyToMessage != nil {
		my = uc.From.ID == message.ReplyToMessage.From.ID
	}
	ip := reIP.FindString(message.Text)
	Data := update.CallbackQuery.Data
	if strings.HasPrefix(Data, "…") {
		ip = ""
	}
	ups := fmt.Sprintf("%s %s @%s #%d%s", uc.From.FirstName, uc.From.LastName, uc.From.Username, uc.From.ID, notAllowed(my, 0, ul))
	bot.AnswerCallbackQuery(mainCtx, &tg.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: ups + tf(ips.count() == 0, "∅", ip+Data), ShowAlert: !my})
	if !my {
		return nil
	}
	if Data == "❎" {
		bot.DeleteMessage(mainCtx, &tg.DeleteMessageParams{ChatID: tu.ID(tm.GetChat().ID), MessageID: tm.GetMessageID()})
		return nil
	}
	if Data == "…" {
		rm := tu.InlineKeyboard(message.ReplyMarkup.InlineKeyboard[0])
		if len(message.ReplyMarkup.InlineKeyboard) == 1 {
			if ips.count() == 0 {
				return nil
			}
			rm = tu.InlineKeyboard(message.ReplyMarkup.InlineKeyboard[0], tu.InlineKeyboardRow(ikbs[:len(ikbs)-1]...))
		}
		bot.EditMessageReplyMarkup(mainCtx, &tg.EditMessageReplyMarkupParams{ChatID: tu.ID(tm.GetChat().ID), MessageID: tm.GetMessageID(), ReplyMarkup: rm})
		return nil
	}

	if ips.count() == 0 {
		return nil
	}
	if strings.HasPrefix(Data, "…") {
		ips.update(customer{Cmd: strings.TrimPrefix(Data, "…")})
	} else {
		ips.write(ip, customer{Cmd: Data})
	}
	return nil
}

// handler DeleteMessage
func bhReplyMessageIsMinus(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	re := update.Message.ReplyToMessage
	err := bot.DeleteMessage(mainCtx, &tg.DeleteMessageParams{ChatID: tu.ID(re.Chat.ID), MessageID: re.MessageID})
	if err != nil {
		let.Println(err)
		bot.EditMessageText(mainCtx, &tg.EditMessageTextParams{ChatID: tu.ID(re.Chat.ID), MessageID: re.MessageID, Text: "-"})
	}
	return nil
}

// send t.C then reset t
func restart(t *time.Ticker, d time.Duration) {
	if t != nil {
		t.Reset(time.Millisecond * 100)
		time.Sleep(time.Millisecond * 150)
		t.Reset(d)
	}
}

// handler Command
func bhAnyCommand(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	tm := update.Message
	if tm == nil {
		return nil
	}
	if tm.Chat.Type == "private" {
		p := "/start "
		if strings.HasPrefix(tm.Text, p) {
			ds, err := base64.StdEncoding.DecodeString(strings.Trim(strings.TrimPrefix(tm.Text, p), " "))
			if err == nil {
				ltf.Println(string(ds))
				tm.Text = p + string(ds)
				switch {
				case reYYYYMMDD.MatchString(tm.Text):
					_ = bhEasterEgg(ctx, update)
				case reIP.MatchString(tm.Text):
					_ = bhAnyWithMatch(ctx, update)
				}
				return nil
			}
		}
		// For owner as first chatID in args
		if tm.From != nil && chats[:1].allowed(tm.From.ID) {
			p = "/restart"
			if strings.HasPrefix(tm.Text, p) {
				restart(tacker, tt)
				return nil
			}
			p = "/stop"
			if strings.HasPrefix(tm.Text, p) {
				closer.Close()
				return nil
			}
		}
	}
	ok, ups := allowed(ul, tm.From.ID, tm.Chat.ID)
	mecs := []tu.MessageEntityCollection{
		tu.Entity(dic.add(ul,
			"en:List of IP addresses expected\n",
			"ru:Ожидался список IP адресов\n",
		)),
		tu.Entity("/127.0.0.1 127.0.0.2 127.0.0.254").Code(),
		tu.Entity(ups),
	}
	mecsf := len(mecs) - 1
	if ok {
		mecsf = 0
	}
	ikbsf = len(ikbs) - 1
	if chats.allowed(tm.From.ID) && ips.count() > 0 {
		ikbsf = 0
	}
	params := tu.MessageWithEntities(tu.ID(tm.Chat.ID), mecs[mecsf:]...)
	params.ReplyParameters = &tg.ReplyParameters{MessageID: tm.MessageID}
	params.ReplyMarkup = tu.InlineKeyboard(tu.InlineKeyboardRow(ikbs[ikbsf:]...))
	_, err := bot.SendMessage(mainCtx, params)
	if err != nil {
		let.Println(err)
	}
	return nil
}

// handler LeftChat
func bhLeftChat(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	tm := update.Message
	params := tu.MessageWithEntities(tu.ID(tm.Chat.ID),
		tu.Entity(dic.add(ul,
			"en:He flew away, but promised to return❗\n    ",
			"ru:Он улетел, но обещал вернуться❗\n    ",
		)),
		tu.Entity(dic.add(ul,
			"en:Cute...",
			"ru:Милый...",
		)).Bold(), tu.Entity("😍\n        "),
		tu.Entity(dic.add(ul,
			"en:Cute...",
		)).Italic(), tu.Entity("😢"),
	)
	params.ReplyParameters = &tg.ReplyParameters{MessageID: tm.MessageID}
	_, err := bot.SendMessage(mainCtx, params)
	if err != nil {
		let.Println(err)
	}
	return nil
}

// handler NewMember
func bhNewMember(ctx *th.Context, update tg.Update) error {
	bot := ctx.Bot()
	tm := update.Message
	if !chats.allowed(tm.Chat.ID) {
		return nil
	}
	for _, nu := range tm.NewChatMembers {
		ltf.Println(nu.ID)
		params := tu.MessageWithEntities(tu.ID(tm.Chat.ID),
			tu.Entity(dic.add(ul,
				"en:Hello villagers!",
				"ru:Здорово, селяне!\n",
			)),
			tu.Entity(dic.add(ul,
				"en:Is carriage ready?\n",
				"ru:Карета готова?\n",
			)).Strikethrough(),
			tu.Entity(dic.add(ul,
				"en:The cart is ready!🏓",
				"ru:Телега готова!🏓",
			)),
		)
		params.ReplyParameters = &tg.ReplyParameters{MessageID: tm.MessageID}
		_, err := bot.SendMessage(mainCtx, params)
		if err != nil {
			let.Println(err)
		}
		break
	}
	return nil
}

// is key in args
func allowed(key string, ChatIDs ...int64) (ok bool, s string) {
	s = "\n🏓"
	for _, v := range ChatIDs {
		ok = chats.allowed(v)
		if ok {
			return
		}
	}
	s = notAllowed(false, ChatIDs[0], key)
	return
}

// message for ChatID
func notAllowed(ok bool, ChatID int64, key string) (s string) {
	s = "\n🏓"
	if ok {
		return
	}
	s = dic.add(key,
		"en:\nNot allowed for you",
		"ru:\nБатюшка не благословляет Вас",
	)
	if ChatID != 0 {
		s += fmt.Sprintf(":%d", ChatID)
	}
	s += "\n🏓"
	return
}

// tm info
func fcRfRc(tm *tg.Message) (s string) {
	s = ""
	if tm == nil {
		return
	}
	s = fmt.Sprintf("From:@%s #%d Chat:@%s #%d", tm.From.Username, tm.From.ID, tm.Chat.Title, tm.Chat.ID)
	if tm.ReplyToMessage == nil {
		return
	}
	s = fmt.Sprintf(" Reply From:@%s #%d Reply Chat:@%s #%d", tm.ReplyToMessage.From.Username, tm.ReplyToMessage.From.ID, tm.ReplyToMessage.Chat.Title, tm.ReplyToMessage.Chat.ID)
	return
}

// encode for /start
func start(me *tg.User, s string) string {
	return fmt.Sprintf("t.me/%s?start=%s", me.Username, base64.StdEncoding.EncodeToString([]byte(s)))
}
