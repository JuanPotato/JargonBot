// bot.go
package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	BotToken = ""
)

var replaceRegex *regexp.Regexp
var JargonBot *tgbotapi.BotAPI

func main() {

	replaceRegex, _ = regexp.Compile("{\\d}")

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Panic(err)
	}

	JargonBot = bot

	//	bot.Debug = true

	username := JargonBot.Self.UserName
	log.Printf("Authorized on account %s", username)

	upConfig := tgbotapi.NewUpdate(0)
	upConfig.Timeout = 60
	techRegex, _ := regexp.Compile(fmt.Sprintf("^\\/tech(?:nology)?(?:@%s)?", username))
	audioRegex, _ := regexp.Compile(fmt.Sprintf("^\\/audio(?:@%s)?", username))
	excuseRegex, _ := regexp.Compile(fmt.Sprintf("^\\/excuse(?:@%s)?", username))
	startRegex, _ := regexp.Compile(fmt.Sprintf("^\\/(start|about)(?:@%s)?", username))
	helpRegex, _ := regexp.Compile(fmt.Sprintf("^\\/help(?:@%s)?", username))

	updates, err := JargonBot.GetUpdatesChan(upConfig)

	for update := range updates {
		if update.InlineQuery != nil {
			go JargonInline(update)
		} else if update.Message != nil {
			switch true {
			case techRegex.MatchString(update.Message.Text):
				go Jargon(0, update)
			case audioRegex.MatchString(update.Message.Text):
				go Jargon(1, update)
			case excuseRegex.MatchString(update.Message.Text):
				go Jargon(2, update)
			case helpRegex.MatchString(update.Message.Text):
				go help(update)
			case startRegex.MatchString(update.Message.Text),
				update.Message.NewChatMember != nil && update.Message.NewChatMember.ID == JargonBot.Self.ID:
				go about(update)
			default:
				if update.Message.Chat.Type == "private" {
					go about(update)
				}
			}
		}
	}
}

func help(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	msg.ReplyToMessageID = update.Message.MessageID

	JargonBot.Send(msg)
}

func about(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, aboutText)
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ParseMode = "HTML"

	JargonBot.Send(msg)
}

func Jargon(kind int, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, Jargen(kind))
	msg.ReplyToMessageID = update.Message.MessageID

	JargonBot.Send(msg)
}

func JargonInline(update tgbotapi.Update) {
	techText := Jargen(0)
	audioText := Jargen(1)
	excuseText := Jargen(2)
	cr := md5.New()
	cr.Write([]byte(techText))
	techArt := tgbotapi.NewInlineQueryResultArticle(string(cr.Sum(nil)), "Technology jargon", techText)
	techArt.Description = techText
	cr.Reset()
	cr.Write([]byte(audioText))
	audioArt := tgbotapi.NewInlineQueryResultArticle(string(cr.Sum(nil)), "Audio jargon", audioText)
	audioArt.Description = audioText
	cr.Reset()
	cr.Write([]byte(excuseText))
	excuseArt := tgbotapi.NewInlineQueryResultArticle(string(cr.Sum(nil)), "Excuse jargon", excuseText)
	excuseArt.Description = excuseText

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{techArt, audioArt, excuseArt},
	}
	if _, err := JargonBot.AnswerInlineQuery(inlineConf); err != nil {
		log.Println(err)
	}
}

func Jargen(kind int) (text string) {
	switch kind {
	case 0:
		text = TechnicalConstructs[rand.Intn(len(TechnicalConstructs))]
		text = replaceRegex.ReplaceAllStringFunc(text, techRepl)
	case 1:
		text = AudioConstructs[rand.Intn(len(AudioWordPool))]
		text = replaceRegex.ReplaceAllStringFunc(text, audioRepl)
	case 2:
		text = ExcuseConstructs[rand.Intn(len(ExcuseWordPool))]
		text = replaceRegex.ReplaceAllStringFunc(text, excuseRepl)
	}
	return
}

func techRepl(s string) string {
	i, _ := strconv.Atoi(s[1:2])
	return TechnicalWordPool[i][rand.Intn(len(TechnicalWordPool))]
}

func audioRepl(s string) string {
	i, _ := strconv.Atoi(s[1:2])
	return AudioWordPool[i][rand.Intn(len(AudioWordPool))]
}

func excuseRepl(s string) string {
	i, _ := strconv.Atoi(s[1:2])
	return ExcuseWordPool[i][rand.Intn(len(ExcuseWordPool))]
}
