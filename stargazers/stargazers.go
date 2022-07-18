package stargazers

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anaskhan96/soup"
)

type Stargazers struct {
	Username     string `json:"username"`
	Follower     int64  `json:"follower"`
	Contribution int64  `json:"contribution"`
}

var STARGAZER_PER_PAGE = 48

func FindStargazers(repo string) ([]string, error) {

	users := []string{}

	resp, err := soup.Get(fmt.Sprintf("%s/stargazers", repo))
	if err != nil {
		return users, err
	}
	doc := soup.HTMLParse(resp)

	if doc.Error != nil {
		fmt.Println(repo)
	}

	links := doc.Find("div", "id", "repos").FindAll("a", "data-hovercard-type", "user")
	for _, link := range links {
		user := link.Text()
		if len(user) > 0 {
			users = append(users, link.Text())
		}
	}

	counter := doc.Find("nav", "class", "tabnav-tabs").Find("span", "class", "Counter")
	numStargazers, _ := convertNumContribution(counter.FullText())

	numPage := int(math.Round(float64(numStargazers) / float64(STARGAZER_PER_PAGE)))

	//Github only allow get 100 page stargazers
	if numPage > 100 {
		numPage = 100
	}

	fmt.Println("numStargazers:", numStargazers)
	// fmt.Println("numPage:", numPage)

	var wg sync.WaitGroup

	//get next page
	for i := 2; i <= numPage; i++ {
		url := fmt.Sprintf("%s/stargazers?page=%d", repo, i)
		wg.Add(1)
		go func(url string) {

			// fmt.Println(url)
			defer wg.Done()

			x, err := getStargazers(url)
			if err != nil {
				return
			}
			users = append(users, x...)

		}(url)
	}
	wg.Wait()
	return users, nil
}

func getStargazers(url string) ([]string, error) {
	users := []string{}

	resp, err := soup.Get(url)
	if err != nil {
		return users, err
	}
	doc := soup.HTMLParse(resp)
	if doc.Error != nil {
		fmt.Println("getStargazers:", url)
	}
	repo := doc.Find("div", "id", "repos")
	if repo.Error == nil {
		links := repo.FindAll("a", "data-hovercard-type", "user")
		for _, link := range links {
			user := link.Text()
			if len(user) > 0 {
				users = append(users, link.Text())
			}
		}
	} else {
		fmt.Println(url)
	}

	return users, nil
}
func FindFollowerAndContribution(user string) (int64, int64, error) {
	var numFollower int64
	var numContribution int64

	// url := fmt.Sprintf("https://github.com/%s?%d=x", user, time.Now().Unix())
	url := fmt.Sprintf("https://github-com.translate.goog/%s?_x_tr_sl=auto&_x_tr_tl=vi&_x_tr_hl=vi&_x_tr_pto=wapp&t=%s", user, user)

	// resp, err := soup.Get(url)
	var htmlContent string

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		htmlDoc, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return numFollower, numContribution, err
		}
		htmlContent = string(htmlDoc)
	} else {
		// fmt.Println("StatusCode: ", resp.StatusCode)

		return numFollower, numContribution, err
	}

	if strings.Contains(htmlContent, "activity is private") || strings.Contains(htmlContent, "Whoa there") {
		return 0, 0, nil
	}

	doc := soup.HTMLParse(htmlContent)
	if doc.Error != nil {
		fmt.Println(url)
		return 0, 0, doc.Error
	}

	// get follower
	follower := doc.Find("a", "href", fmt.Sprintf("https://github-com.translate.goog/%s?tab=followers&_x_tr_sl=auto&_x_tr_tl=vi&_x_tr_hl=vi&_x_tr_pto=wapp", user))
	if follower.Error == nil {
		follower = follower.Find("span")
		numFollower = convertNumFollower(follower.FullText())
	}

	//get contribution
	contribution := doc.Find("div", "class", "js-yearly-contributions")
	if contribution.Error == nil {
		contribution = contribution.Find("h2")
		// fmt.Println("contribution.FullText()", contribution.FullText())

		s := strings.Split(strings.TrimSpace(contribution.FullText()), " ")
		numContribution, _ = convertNumContribution(strings.TrimSpace(s[0]))
	}

	return numFollower, numContribution, nil
}

func GetStargazers(repo string) ([]Stargazers, error) {

	stargazers := []Stargazers{}
	users, err := FindStargazers(repo)
	if err != nil {
		return stargazers, err
	}
	var wg sync.WaitGroup

	for i, user := range users {
		if i%500 == 0 {
			time.Sleep(2 * time.Second)
		}
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			num1, num2, err := FindFollowerAndContribution(u)
			if err != nil {
				return
			}
			// fmt.Printf("%s follower:%d contribution:%d\n", u, num1, num2)
			stargazers = append(stargazers, Stargazers{Username: u, Follower: num1, Contribution: num2})
		}(user)
	}
	wg.Wait()

	//sort by Follower
	sort.SliceStable(stargazers, func(i, j int) bool {
		return stargazers[i].Follower >= stargazers[j].Follower
	})

	return stargazers, nil
}

func convertNumContribution(text string) (int64, error) {
	replacer := strings.NewReplacer(",", "", "+", "")

	s := replacer.Replace(text)
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func convertNumFollower(text string) int64 {
	num, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		if strings.Contains(text, ".") {
			s := strings.Split(strings.Trim(text, "k"), ".")
			d1, _ := strconv.ParseInt(s[0], 10, 64)
			d2, _ := strconv.ParseInt(s[1], 10, 64)
			return d1*1000 + d2*100
		} else {
			s := strings.Trim(text, "k")
			d1, _ := strconv.ParseInt(s, 10, 64)
			return d1 * 1000
		}
	}
	return num
}
