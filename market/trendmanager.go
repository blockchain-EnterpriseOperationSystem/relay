package market

import (
	"sync"
	"github.com/patrickmn/go-cache"
	"github.com/Loopring/relay/dao"
	"errors"
	"sort"
	"github.com/robfig/cron"
	"log"
	"time"
	"fmt"
)

const (
	OneHour = "1Hr"
	//TwoHour = "2Hr"
	//OneDay = "1Day"
)

type Ticker struct {
	Market	              string
	Interval			  string
	Amount                float64
	Vol                   float64
	Open                  float64
	Close                 float64
	High                  float64
	Low                   float64
	Last 				  float64
	Change				  string
}

type Cache struct {
	trends []Trend
	fills []dao.FillEvent
}

type Trend struct {
	Interval   string
	Market     string
	Vol        float64
	Amount     float64
	CreateTime int64
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Start      int64
	End        int64
}

type TrendManager struct {
	c             *cache.Cache
	IsTickerReady bool
	IsTrendReady  bool
	cacheReady    bool
	rds           dao.RdsService
	cron		  *cron.Cron
}

var once sync.Once
var trendManager TrendManager

const trendKey = "market_ticker"
const tickerKey = "market_ticker_view"

func NewTrendManager(dao dao.RdsService) TrendManager {

	once.Do(func () {
		trendManager = TrendManager{rds:dao, cron:cron.New()}
		trendManager.c = cache.New(cache.NoExpiration, cache.NoExpiration)
		trendManager.initCache()
		//trendManager.startScheduleUpdate()
	})

	return trendManager
}

// ======> init cache steps
// step.1 init all market
// step.2 get all trend record into cache
// step.3 get all order fillFilled into cache
// step.4 calculate 24hr ticker
// step.5 send channel cache ready
// step.6 start schedule update

func (t *TrendManager) initCache() {

	trendMap := make(map[string]Cache)
	tickerMap := make(map[string]Ticker)
	for _, mkt := range AllMarkets {
		mktCache := Cache{}
		mktCache.trends = make([]Trend, 0)
		mktCache.fills = make([]dao.FillEvent, 0)

		// default 100 records load first time
		trends, err := t.rds.TrendPageQuery(dao.Trend{Market:mkt}, 1, 100)

		if err != nil {
			log.Fatal(err)
		}

		for _, trend := range trends.Data {
			mktCache.trends = append(mktCache.trends, dao.ConvertUp(trend.(dao.Trend)))
		}

		tokenS, tokenB, _ := UnWrap(mkt)
		now := time.Now()
		firstSecondThisHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		sbFills, err := t.rds.QueryRecentFills(tokenS, tokenB, firstSecondThisHour.Unix(), 0)
		if err != nil {
			log.Fatal(err)
		}

		bsFills, err := t.rds.QueryRecentFills(tokenB, tokenS, firstSecondThisHour.Unix(), 0)

		if err != nil {
			log.Fatal(err)
		}

		for _, f := range sbFills {
			mktCache.fills = append(mktCache.fills, f)
		}
		for _, f := range bsFills {
			mktCache.fills = append(mktCache.fills, f)
		}

		trendMap[mkt] = mktCache

		ticker := calculateTicker(mkt, mktCache.fills, mktCache.trends, firstSecondThisHour)
		tickerMap[mkt] = ticker
	}
	t.c.Set(trendKey, trendMap, cache.NoExpiration)
	t.c.Set(tickerKey, tickerMap, cache.NoExpiration)

	t.cacheReady = true
	t.startScheduleUpdate()

}

func calculateTicker(market string, fills []dao.FillEvent, trends [] Trend, now time.Time) Ticker {

	var result = Ticker{Market:market}

	before24Hour := now.Unix() - 24 * 60 * 60

	var (
		high float64
		low float64
		vol float64
		amount float64
	)

	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Start < trends[j].Start
	})


	for _, data := range trends {

		if data.Start > before24Hour {
			continue
		}

		vol += data.Vol
		amount += data.Amount
		if high == 0 || high < data.High {
			high = data.High
		}
		if low == 0 || low < data.Low {
			low = data.Low
		}
	}

	for i, data := range fills {

		if IsBuy(data.TokenS) {
			vol += ByteToFloat(data.AmountB)
			amount += ByteToFloat(data.AmountS)
		} else {
			vol += ByteToFloat(data.AmountS)
			amount += ByteToFloat(data.AmountB)
		}

		price := CalculatePrice(data.AmountS, data.AmountB, data.TokenS, data.TokenB)

		if i == len(fills) - 1 {
			result.Last = price
		}

		if high == 0 || high < price {
			high = price
		}
		if low == 0 || low < price {
			low = price
		}
	}

	result.High = high
	result.Low = low

	result.Open = trends[0].Open
	result.Close = trends[len(trends) - 1].Close
	result.Change = fmt.Sprintf("%.2f%%", 100 * result.Last / result.Open)

	result.Vol = vol
	result.Amount = amount
	return result
}

func (t *TrendManager) startScheduleUpdate() {
	t.cron.AddFunc("10 * * * * *", t.insertTrend)
	t.cron.Start()
}


func (t *TrendManager) insertTrend() {
	// get latest 24 hour trend if not exist generate

	for _, mkt := range AllMarkets {
		now := time.Now()
		firstSecondThisHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)

		for i := 1; i < 10; i++ {

			start := firstSecondThisHour.Unix() - int64(i * 60 * 60)
			end := firstSecondThisHour.Unix() - int64((i - 1) * 60 * 60)

			trends, err := t.rds.TrendQueryByTime(mkt, start, end)
			if err != nil {
				log.Println("query trend err", err)
				return
			}

			tokenS, tokenB, _ := UnWrap(mkt)
			if trends == nil || len(trends) == 0  {
				fills , _ := t.rds.QueryRecentFills(tokenS, tokenB, start, end)

				toInsert := dao.Trend{
					Interval:OneHour,
					Market:mkt,
					CreateTime:time.Now().Unix(),
					Start: start,
					End: end}

				var (
					high float64
					low float64
					vol float64
					amount float64
				)

				sort.Slice(fills, func(i, j int) bool {
					return fills[i].CreateTime < fills[j].CreateTime
				})

				for _, data := range fills {

					if IsBuy(data.TokenS) {
						vol += ByteToFloat(data.AmountB)
						amount += ByteToFloat(data.AmountS)
					} else {
						vol += ByteToFloat(data.AmountS)
						amount += ByteToFloat(data.AmountB)
					}

					price := CalculatePrice(data.AmountS, data.AmountB, data.TokenS, data.TokenB)

					if high == 0 || high < price {
						high = price
					}
					if low == 0 || low < price {
						low = price
					}
				}

				toInsert.High = high
				toInsert.Low = low

				openFill := fills[0]
				toInsert.Open = CalculatePrice(openFill.AmountS, openFill.AmountB, openFill.TokenS, openFill.TokenB)
				closeFill := fills[len(fills) - 1]
				toInsert.Close = CalculatePrice(closeFill.AmountS, closeFill.AmountB, closeFill.TokenS, closeFill.TokenB)

				toInsert.Vol = vol
				toInsert.Amount = amount

				if err := t.rds.Add(toInsert); err != nil {
					fmt.Println(err)
				}
			}
		}

	}
}

func(t *TrendManager) GetTrends(market string) (trends []Trend, err error) {

	if t.cacheReady {
		if trendCache, ok := t.c.Get(trendKey); !ok {
			err = errors.New("can't found trends by key : " + trendKey)
		} else {
			trends = trendCache.(map[string][]Trend)[market]
		}
	} else {
		err = errors.New("cache is not ready , please access later")
	}
	return
}

func(t *TrendManager) GetTicker() (tickers [] Ticker, err error) {

	if t.cacheReady {
		if tickerInCache, ok := t.c.Get(tickerKey); ok {
			tickerMap := tickerInCache.(map[string]Ticker)
			tickers = make([]Ticker, len(tickerMap))
			for _, v := range tickerMap {
				tickers = append(tickers, v)
			}
		} else {
			err = errors.New("get ticker from cache error, no value found")
		}
	} else {
		err = errors.New("cache is not ready , please access later")
	}
	return
}
