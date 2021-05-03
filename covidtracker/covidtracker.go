package covidtracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/dustin/go-humanize"
	"github.com/tidwall/gjson"
)

const (
	TIMEOUT = 10 // second
	PERCENT = 0.01
)

// FetchPopulation looks up the population of country from wikidata.
// In case of error, the population is set to -1.
// Note: country selection is not yet implemented, it's hard coded to be Hungary at the moment.
func FetchPopulation(country string) (int, error) {
	// wikidata endpoint
	endpoint := "https://query.wikidata.org/sparql"

	query := "SELECT ?population WHERE { " +
		"?country wdt:P31 wd:Q6256." +
		"?country wdt:P17 wd:Q28." +
		"?country wdt:P1082 ?population." +
		"SERVICE wikibase:label { bd:serviceParam wikibase:language \"en\". }}"

	form := url.Values{}
	form.Set("format", "json")
	form.Set("query", query)
	reqBody := bytes.NewBuffer([]byte(form.Encode()))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, reqBody)
	if err != nil {
		return -1, fmt.Errorf("cannot create http request to wikibase | %w", err)
	}

	req.Header.Add("User-Agent", "covidtracker")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("cannot fetch wikibase response | %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -11, fmt.Errorf("cannot read body of wikibase response | %w", err)
	}

	bodyJson := string(body)

	num := gjson.Get(bodyJson, "results.bindings.0.population.value").Int()
	fmt.Println(num)

	return int(num), nil
}

// drawAsciiProgressBar draws an ASCII progressbar based on the given percentage and target width.
func drawAsciiProgressBar(percentage float64, width int) string {
	progressBarStr := ""

	markerEmpty := "░"
	markerFull  := "▓"

	fullCount := int(math.Round(float64(width) * percentage * PERCENT))

	for i := 0; i < fullCount; i++ {
		progressBarStr += markerFull
	}
	for i := fullCount+1; i <= width; i++ {
		progressBarStr += markerEmpty
	}

	return progressBarStr
}

// FetchCVNLast fetches the last reposted vaccination numbers from the last Twitter status update.
// Unfortunately in Hungary there is no official API and since this bot aims to be stateless and minimal,
// it's much easier to get the last tweet than to keep a database for just this, although that would be a proper design
// choice for anything bigger or more complex.
func FetchCVNLast() (int, error) {
	// Twitter client
	client, errTwitterClient := NewGoTwitterClient()
	if errTwitterClient != nil {
		return -1, fmt.Errorf("cannot create Go-Twitter client | %w", errTwitterClient)
	}

	// get timeline
	tweets, resp, err := client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
		Count: 1,
	})

	if resp.StatusCode != http.StatusOK || err != nil {
		fmt.Println("aww :(", resp.StatusCode)
		return -1, err
	}

	// fetch last vaccination number
	fields := strings.Split(tweets[0].Text, "|")
	numStr := fields[len(fields)-1]
	numStr = strings.ReplaceAll(numStr, " ", "")

	if len(numStr) == 0 {
		return -1, errors.New("empty last tweet")
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return -1, err
	}

	return num, nil
}

// FetchCVN parses latest vaccination numbers from official website.
func FetchCVN() (int, error) {
	urlHU := "https://koronavirus.gov.hu/"

	// client
	client := &http.Client{ // nolint:exhaustivestruct
		Timeout: time.Duration(TIMEOUT) * time.Second,
	}

	// make request
	response, err := client.Get(urlHU)
	if err != nil {
		return -1, fmt.Errorf("cannot fetch latest vaccination numbers from website | %w", err)
	}
	defer response.Body.Close()

	// get the response body as a string
	dataInBytes, err := ioutil.ReadAll(response.Body)
	pageContent := string(dataInBytes)

	// parse vaccination number
	numRegexp := regexp.MustCompile(`<div id="api-beoltottak">(.*)</div>`)
	numRegexpMatchList := numRegexp.FindAllStringSubmatch(pageContent, -1)
	if len(numRegexpMatchList[0]) == 2 {
		numStr := numRegexpMatchList[0][1]
		numStr = strings.ReplaceAll(numStr, " ", "")

		// convert number string to integer
		if num, err := strconv.Atoi(numStr); err == nil {
			return num, nil
		} else {
			fmt.Errorf("cannot convert vaccination string count to int | %w", err)
		}
	}

	return -1, fmt.Errorf("cannot convert vaccination string count to int")
}

// Tweet sends a Twitter status update containing the latest vaccination statistics.
func Tweet(vaccinatedNum int, population int) error {
	// calculate vaccination progress
	percentage := float64(vaccinatedNum) / float64(population) / PERCENT
	progressBarWidth := 20

	// assemble message
	message := fmt.Sprintf("%s | %.2f%% | %s",
		drawAsciiProgressBar(percentage, progressBarWidth),
		percentage,
		humanize.FormatInteger("# ###.", vaccinatedNum))

	// fmt.Println(message)

	// Twitter client
	client, errTwitterClient := NewGoTwitterClient()
	if errTwitterClient != nil {
		return fmt.Errorf("cannot create Go-Twitter client | %w", errTwitterClient)
	}

	// Send a Tweet
	_, _, errTweet := client.Statuses.Update(message, nil)

	if errTweet != nil {
		return fmt.Errorf("could not send tweet | %w", errTweet)
	}

	// fmt.Println(tweet, resp)

	return nil
}

// TwitterAuth represents all the keys and secrets that is needed for oath1.
type TwitterAuth struct {
	apiKey            string
	apiSecretKey      string
	accessToken       string
	accessTokenSecret string
}

// getTwitterAuth fetched a TwitterAuth object from os env.
func getTwitterAuth() (TwitterAuth, error) {
	twitterAuth := TwitterAuth {
		apiKey: os.Getenv("API_KEY"),
		apiSecretKey: os.Getenv("API_SECRET_KEY"),
		accessToken: os.Getenv("ACCESS_TOKEN"),
		accessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
	}

	if len(twitterAuth.apiKey) == 0 {
		return twitterAuth, fmt.Errorf("missing API KEY")
	}
	if len(twitterAuth.apiSecretKey) == 0 {
		return twitterAuth, fmt.Errorf("missing API Secret Key")
	}
	if len(twitterAuth.accessToken) == 0 {
		return twitterAuth, fmt.Errorf("missing API Access Token")
	}
	if len(twitterAuth.accessTokenSecret) == 0 {
		return twitterAuth, fmt.Errorf("missing API Access Token Secret")
	}

	return twitterAuth, nil
}

// NewGoTwitterClient helper function creating the new Go-Twitter client.
func NewGoTwitterClient() (*twitter.Client, error){
	twitterAuth, err := getTwitterAuth()
	if err != nil {
		return nil, fmt.Errorf("cannot get twitter auth, %w", err)
	}

	config := oauth1.NewConfig(twitterAuth.apiKey, twitterAuth.apiSecretKey)
	token := oauth1.NewToken(twitterAuth.accessToken, twitterAuth.accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	httpClient.Timeout = time.Duration(TIMEOUT) * time.Second

	// Twitter client
	client := twitter.NewClient(httpClient)

	return client, nil
}
