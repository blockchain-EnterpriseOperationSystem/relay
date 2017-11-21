package market

import (
	"math/big"
	"errors"
	"strings"
	"fmt"
)

const weiToEther = 1e18

func ByteToFloat(amount [] byte) float64 {
	var rst big.Int
	rst.UnmarshalText(amount)
	return float64(rst.Int64())/weiToEther
}

func FloatToByte(amount float64) [] byte {
	rst, _ := big.NewInt(int64(amount * weiToEther)).MarshalText()
	return rst
}

var SupportTokens = map[string]string {
	"lrc" : "0xskdfjdkfj",
	"coss" : "0xskdjfskdfj",
}

var SupportMarket = map[string]string {
	"weth" : "0xsldkfjsdkfj",
}

var AllTokens = func() map[string]string {
	all := make(map[string]string)
	for k, v := range SupportMarket {
		all[k] = v
	}
	for k, v := range SupportTokens {
		all[k] = v
	}
	return all
}()

var AllMarkets = AllMarket()

func WrapMarket(s, b string) (market string, err error) {

	s, b = strings.ToLower(s), strings.ToLower(b)

	if SupportMarket[s] != "" &&  SupportTokens[b] != "" {
		market = fmt.Sprintf("%s-%s", b, s)
	} else if SupportMarket[b] != "" &&  SupportTokens[s] != "" {
		market = fmt.Sprintf("%s-%s", s, b)
	} else {
		err = errors.New(fmt.Sprintf("not supported market type : %s-%s", s, b))
	}
	return
}

func WrapMarketByAddress(s, b string) (market string, err error) {
	return WrapMarket(AddressToAlias(s), AddressToAlias(b))
}

func UnWrap(market string) (s, b string, err error) {
	mkt := strings.Split(strings.TrimSpace(market), "-")
	if len(mkt) != 2 {
		err = errors.New("only support market format like tokenS-tokenB")
		return
	}

	s, b = strings.ToLower(mkt[0]), strings.ToLower(mkt[1])
	return
}

func IsSupportedToken(token string) bool {
	return SupportTokens[token] != ""
}

func AliasToAddress(t string) string {
	return AllTokens[t]
}

func AddressToAlias(t string) string {
	for k, v := range AllTokens {
		if t == v {
			return k
		}
	}
	return ""
}

func AllMarket() []string {
	mkts := make([]string, 0)
	for k := range SupportTokens {
		for kk := range SupportMarket {
			mkts = append(mkts, k + "-" + kk)
		}
	}
	return mkts
}

func CalculatePrice(amountS, amountB []byte, s, b string) float64 {

	as := ByteToFloat(amountS)
	ab := ByteToFloat(amountB)

	if as == 0 || ab == 0 {
		return 0
	}

	if IsBuy(s) {
		return ab/as
	}

	return as/ab

}

func IsBuy(s string) (bool) {
	if IsAddress(s) {
		s = AddressToAlias(s)
	}
	if SupportTokens[s] != "" {
		return false
	}
	return true
}

func IsAddress(token string) bool {
	return strings.HasPrefix(token, "0x")
}
