package slack

import (
	"log"
	"os"

	"github.com/jobandtalent/hal"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (a *adapter) startConnection() {
	api := slack.New(
		a.token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(a.botToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	users, err := api.GetUsers()
	if err != nil {
		hal.Logger.Debugf("%s\n", err)
	}

	for _, user := range users {
		// retrieve the name and mention name of our bot from the server
		// if user.Id == api.Id {
		// 	a.name = user.Name
		// 	// skip adding the bot to the users map
		// 	continue
		// }
		// Initialize a newUser object in case we need it.
		newUser := hal.User{
			ID:   user.ID,
			Name: user.Name,
		}
		// Prepopulate our users map because we can easily do so.
		// If a user doesn't exist, set it.
		u, err := a.Robot.Users.Get(user.ID)
		if err != nil {
			a.Robot.Users.Set(user.ID, newUser)
		}

		// If the user doesn't match completely (say, if someone changes their name),
		// then adjust what we have stored.
		if u.Name != user.Name {
			a.Robot.Users.Set(user.ID, newUser)
		}
	}
	hal.Logger.Debugf("Stored users: %s\n", a.Robot.Users.All())
	hal.Logger.Info("Retrieved slack users")

	a.socketmode = client
	go func() {
		for evt := range a.socketmode.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				hal.Logger.Debug("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				hal.Logger.Debug("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				hal.Logger.Debug("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					hal.Logger.Debugf("Ignored %+v\n", evt)

					continue
				}
				a.socketmode.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						hal.Logger.Debugf("Received message: %+v", ev)
						msg := a.newMessage(ev)
						a.Receive(msg)
					case *slackevents.TeamJoinEvent:
						hal.Logger.Debugf("New member joined the team: %v", ev.User)
						if _, err := a.Robot.Users.Get(ev.User.ID); err != nil {
							a.Robot.Users.Set(ev.User.ID, hal.User{ID: ev.User.ID, Name: ev.User.Name})
						}
					default:
						hal.Logger.Debugf("Unsupported Events API event received: %+v", ev)
					}
				default:
					hal.Logger.Debugf("Unsupported event type received: %s", evt.Type)
				}
			}
		}
	}()

	a.socketmode.Run()
}

func (a *adapter) newMessage(msg *slackevents.MessageEvent) *hal.Message {
	user, _ := a.Robot.Users.Get(msg.User)
	return &hal.Message{
		User: user,
		Room: msg.Channel,
		Text: msg.Text,
	}
}
