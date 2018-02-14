package metacrawl

import (
	"bytes"
	"encoding/csv"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const (
	// TaskNotStarted is a custom task status
	TaskNotStarted = "not started"
	// TaskInProgress is a custom task status
	TaskInProgress = "in progress"
	// TaskCompleted is a custom task status
	TaskCompleted = "completed"
)

// MCTask is a MetaCrawl task implemetation
type MCTask struct {
	status           string
	statusMutex      *sync.RWMutex
	timeout          time.Duration
	httpClient       *http.Client
	resultBuffer     *bytes.Buffer
	resultMutex      *sync.Mutex
	metaCrawlSvc     Svc
	metaAttrRegistry map[string][]string
	urls             []string
}

// NewMetaCrawlTask is a MetaCrawlTask constructor
func NewMetaCrawlTask(
	mCrawl Svc,
	urls []string,
	timeout time.Duration,
) *MCTask {
	// Each timeout has the same value
	// TODO: allow to specify individual timeouts
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout,
	}
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: netTransport,
	}

	// define desired meta attributes
	metaAttributes := map[string][]string{
		"name":     []string{"description", "keywords"},
		"property": []string{"og:image"},
	}

	return &MCTask{
		metaCrawlSvc:     mCrawl,
		urls:             urls,
		httpClient:       httpClient,
		resultBuffer:     bytes.NewBuffer(nil),
		resultMutex:      &sync.Mutex{},
		metaAttrRegistry: metaAttributes,
		status:           TaskNotStarted,
		statusMutex:      &sync.RWMutex{},
	}
}

func extractMetaContent(metaToken html.Token) string {
	for _, attr := range metaToken.Attr {
		if attr.Key == "content" {
			return strings.TrimSpace(attr.Val)
		}
	}

	return ""
}

func parseMetaAttributes(
	metaToken html.Token,
	attrRegistry map[string][]string,
) (metaKey, metaValue string) {
	for _, attr := range metaToken.Attr {
		if _, ok := attrRegistry[attr.Key]; ok {
			for _, attrValue := range attrRegistry[attr.Key] {
				if attr.Val == attrValue {
					return attrValue, extractMetaContent(metaToken)
				}
			}
		}
	}

	return "", ""
}

func (mt *MCTask) parseMetaTags(body io.Reader) []string {
	tokenizer := html.NewTokenizer(body)
	var title string
	metaResult := make(map[string]string)
	complete := false
	for {
		if !complete {
			tokenType := tokenizer.Next()
			switch tokenType {
			case html.ErrorToken:
				complete = true
				break
			case html.StartTagToken, html.SelfClosingTagToken:
				startTagToken := tokenizer.Token()
				if startTagToken.Data == "head" {
					for {
						if !complete {
							headTokenType := tokenizer.Next()
							headTagToken := tokenizer.Token()

							switch headTokenType {
							case html.StartTagToken, html.SelfClosingTagToken:
								// get page title
								if headTagToken.Data == "title" {
									tt := tokenizer.Next()

									if tt == html.TextToken {
										next := tokenizer.Token()
										title = strings.TrimSpace(next.Data)
									}
								}
								// get meta tag values
								if headTagToken.Data == "meta" {
									mkey, mvalue := parseMetaAttributes(headTagToken, mt.metaAttrRegistry)
									metaResult[mkey] = mvalue
								}
							case html.EndTagToken:
								if headTagToken.Data == "head" {
									complete = true
									break
								}
							}
						} else {
							return []string{title, metaResult["description"], metaResult["keywords"], metaResult["og:image"]}
						}
					}
				}
			}
		}
	}
}

func (mt *MCTask) appendToResult(csvWriter *csv.Writer, row []string) {
	mt.resultMutex.Lock()
	csvWriter.Write(row)
	mt.resultMutex.Unlock()
}

func (mt *MCTask) processURL(csvWriter *csv.Writer, rawURL string) {
	logger := mt.metaCrawlSvc.Logger()
	urlOk := govalidator.IsURL(rawURL) //url.ParseRequestURI(rawURL)
	if !urlOk {
		mt.appendToResult(csvWriter, []string{"-1", rawURL, "", "", "", ""})
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	rLimitter := mt.metaCrawlSvc.RateLimitterForDomain(u.Host)
	<-rLimitter.C
	logger.Debug("crawling", zap.String("url", rawURL))

	resp, err := mt.httpClient.Get(rawURL)
	if err != nil {
		mt.appendToResult(csvWriter, []string{"0", rawURL, "", "", "", ""})
		return
	}

	body := resp.Body
	defer body.Close()

	// convert body to utf-8 charset
	utf8Body, err := charset.NewReader(body, resp.Header.Get("Content-Type"))
	if err != nil {
		return
	}

	record := []string{strconv.Itoa(resp.StatusCode), rawURL}
	// parse meta tags from body
	metaTags := mt.parseMetaTags(utf8Body)

	// append them to the result slice
	record = append(record, metaTags...)
	mt.appendToResult(csvWriter, record)
}

// Process is a MetaCrawl task process method that does all the heavy-lifting
func (mt *MCTask) Process() error {
	csvWriter := csv.NewWriter(mt.resultBuffer)
	// write csv header
	mt.appendToResult(csvWriter, []string{"HTTP Status Code", "URL", "Page Title", "Meta Description", "Meta Keywords", "Og:image"})

	mt.setStatus(TaskInProgress)

	var wg sync.WaitGroup
	wg.Add(len(mt.urls))
	for _, rawURL := range mt.urls {
		go func(rURL string) {
			defer wg.Done()
			mt.processURL(csvWriter, rURL)
		}(rawURL)
	}
	wg.Wait()
	csvWriter.Flush()

	mt.setStatus(TaskCompleted)

	return nil
}

func (mt *MCTask) setStatus(status string) {
	mt.statusMutex.Lock()
	mt.status = status
	mt.statusMutex.Unlock()
}

// Render returns csv data from result buffer
func (mt *MCTask) Render() []byte {
	mt.resultMutex.Lock()
	resultBytes := mt.resultBuffer.Bytes()
	mt.resultMutex.Unlock()
	return resultBytes
}

// Status returns information about task status
func (mt *MCTask) Status() string {
	mt.statusMutex.RLock()
	status := mt.status
	mt.statusMutex.RUnlock()
	return status
}

// Task is an interface describing MetaCrawl task
type Task interface {
	Status() string
	Process() error
	Render() []byte
}
