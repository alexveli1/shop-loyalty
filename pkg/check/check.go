package check

import (
	"fmt"
	"strconv"

	mylog "github.com/alexveli/diploma/pkg/log"
)

func CheckOrderNumber(orderid int64) bool {
	fullStringOrder := fmt.Sprint(orderid)
	stringOrderWithoutCheckNum := fullStringOrder[0 : len(fullStringOrder)-1]
	cleanOrderId, err := strconv.ParseInt(stringOrderWithoutCheckNum, 10, 64)
	if err != nil {
		mylog.SugarLogger.Errorf("error converting clean order number to int64, %v", err)

		return false
	}
	stringCheckNumber := fullStringOrder[len(fullStringOrder)-1 : len(fullStringOrder)]
	intCheckNumber, err := strconv.ParseInt(stringCheckNumber, 10, 64)
	if err != nil {
		mylog.SugarLogger.Errorf("error converting check number to int64, %v", err)

		return false
	}
	if CalculateLuhn(cleanOrderId) == intCheckNumber {

		return true
	}

	return false
}

func CalculateLuhn(number int64) int64 {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return 10 - checkNumber
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
