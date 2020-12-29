package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jfyne/live"
)

const (
	tick = "tick"
)

type clock struct {
	Time time.Time
}

func newClock(s *live.Socket) *clock {
	c, ok := s.Data.(*clock)
	if !ok {
		return &clock{
			Time: time.Now(),
		}
	}
	return c
}

func (c clock) FormattedTime() string {
	return c.Time.Format("15:04:05")
}

func mount(ctx context.Context, view *live.View, r *http.Request, s *live.Socket, connected bool) (interface{}, error) {
	// Take the socket data and tranform it into our view model if it is
	// available.
	c := newClock(s)

	// If we are mouting the websocket connection, trigger the first tick
	// event.
	if connected {
		go func() {
			time.Sleep(1 * time.Second)
			view.Self(s, live.Event{T: tick})
		}()
	}
	return c, nil
}

func main() {
	cookieStore := sessions.NewCookieStore([]byte("weak-secret"))
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = true
	cookieStore.Options.SameSite = http.SameSiteStrictMode

	t, err := template.ParseFiles("examples/root.html", "examples/clock/view.html")
	if err != nil {
		log.Fatal(err)
	}

	view, err := live.NewView(t, "session-key", cookieStore)
	if err != nil {
		log.Fatal(err)
	}

	// Set the mount function for this view.
	view.Mount = mount

	// Server side events.

	// tick event updates the clock every second.
	view.HandleSelf(tick, func(s *live.Socket, _ map[string]interface{}) (interface{}, error) {
		// Get our view model
		c := newClock(s)
		// Update the time.
		c.Time = time.Now()
		// Send ourselves another tick in a second.
		go func() {
			time.Sleep(1 * time.Second)
			view.Self(s, live.Event{T: tick})
		}()
		return c, nil
	})

	// Run the server.
	http.Handle("/clock", view)
	http.Handle("/live.js", live.Javascript{})
	http.Handle("/auto.js.map", live.JavascriptMap{})
	http.ListenAndServe(":8080", nil)
}
