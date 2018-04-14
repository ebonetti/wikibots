package wikibots

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ebonetti/wikibots/internal/botnames"
	"github.com/pkg/errors"
)

//New returns a map ID to name of Wikipedia bots
func New(ctx context.Context, lang string) (ID2bot map[uint32]string, err error) {
	bots, err := botnames.New()
	if err != nil {
		return
	}

	return users(ctx, bots, lang)
}

func users(ctx context.Context, usernames []string, lang string) (userID2User map[uint32]string, err error) {
	userID2User = make(map[uint32]string, len(usernames))

	for _, query := range toURL(usernames, lang) {
		var ud usersData
		for t := time.Second; t < time.Hour; t *= 2 { //exponential backoff
			ud, err = usersDataFrom(ctx, query)
			if err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return nil, errors.Wrap(ctx.Err(), "Error: change in context state")
			case <-time.After(t):
				//do nothing
			}
		}

		if err != nil {
			userID2User = nil
			break
		}

		for _, u := range ud.Query.Users {
			if u.Missing {
				continue
			}
			userID2User[u.UserID] = u.Name
		}
	}

	return
}

func toURL(names []string, lang string) (URLs []string) {
	for _, names := range chunkerize(names) {
		URLs = append(URLs, fmt.Sprintf(base, lang, url.QueryEscape(strings.Join(names, "|"))))
	}
	return
}

const uslimit = 50
const base = "https://%v.wikipedia.org/w/api.php?action=query&list=users&format=json&formatversion=2&ususers=%v"

func chunkerize(names []string) (chunks [][]string) {
	N := (len(names) + uslimit - 1) / uslimit
	for i := 0; i < N; i++ {
		b := i * uslimit
		e := b + uslimit
		if e > len(names) {
			e = len(names)
		}
		chunks = append(chunks, names[b:e])
	}
	return
}

type usersData struct {
	Batchcomplete interface{}
	Warnings      interface{}
	Query         struct {
		Users []mayMissingUser
	}
}

func usersDataFrom(ctx context.Context, query string) (pd usersData, err error) {
	fail := func(e error) (usersData, error) {
		pd, err = usersData{}, errors.Wrapf(e, "BotId2Name: error with the following query: %v", query)
		return pd, err
	}

	bodyR, err := stream(ctx, query)
	if err != nil {
		return fail(err)
	}
	defer bodyR.Close()

	body, err := ioutil.ReadAll(bodyR)
	if err != nil {
		return fail(err)
	}

	err = json.Unmarshal(body, &pd)
	if err != nil {
		return fail(err)
	}

	if pd.Batchcomplete == nil {
		return fail(errors.Errorf("BotId2Name: incomplete batch with the following query: %v", query))
	}

	if pd.Warnings != nil {
		return fail(errors.Errorf("BotId2Name: warnings - %v - with the following query: %v", pd.Warnings, query))
	}

	return
}

type mayMissingUser struct {
	UserID  uint32
	Name    string
	Missing bool
}

func stream(ctx context.Context, query string) (r io.ReadCloser, err error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		err = errors.Wrap(err, "Error: unable create a request with the following url: "+query)
		return
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		err = errors.Wrap(err, "Error: unable do a request with the following url: "+query)
		return
	}

	r = resp.Body
	return
}

var client = &http.Client{Timeout: time.Minute}
