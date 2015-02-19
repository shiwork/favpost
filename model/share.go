package model

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/shiwork/slack"
)

type Share struct {
	Id        int64
	UserId    int64
	ServiceId int64
}

type Service struct {
	ServiceId int64
	Name      string
}

type Sharing interface {
	Share(anaconda.Tweet) error
}

type Tweet struct {
	*anaconda.Tweet
}

// added tweet function
func (t *Tweet) URL() string {
	return "http://twitter.com/" + t.User.ScreenName + "/status/" + t.IdStr
}

type SlackShare struct {
	WebHookURL string
}

func (s SlackShare) Share(atweet anaconda.Tweet) error {
	tweet := Tweet{&atweet}
	incoming := slack.Incoming{WebHookURL: s.WebHookURL}
	return incoming.Post(
		slack.Payload{
			Attachments: []slack.Attachment{
				slack.Attachment{
					Pretext:  tweet.URL(),
					Title:    tweet.User.Name + " @" + tweet.User.ScreenName,
					Text:     tweet.Text,
					ImageUrl: tweet.Entities.Media[0].Media_url,
				},
			},
		},
	)
}
