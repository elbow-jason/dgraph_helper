package prompt

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
)

var groupsRegex = regexp.MustCompile("^(\\d+(,\\d+)?(-\\d+)?)+$")

// ZeroToOneOnly .
func ZeroToOneOnly(answer interface{}) error {
	answerStr := answer.(string)
	num, err := float64Parser(answerStr)
	if err != nil {
		return err
	}
	if num > 1.0 {
		return fmt.Errorf("Answer must be less than 1.0. Got %.2f", num)
	}
	if num < 0.0 {
		return fmt.Errorf("Answer must be greater than 0.0. Got %.2f", num)
	}
	return nil
}

func float64Parser(numStr string) (float64, error) {
	return strconv.ParseFloat(numStr, 64)
}

// AtLeast1025 .
func AtLeast1025(answer interface{}) error {
	answerStr := answer.(string)
	num, err := float64Parser(answerStr)
	if err != nil {
		return err
	}
	if num < 1025.00 {
		return fmt.Errorf("Answer must be at least 1025.00. Got %.2f", num)
	}
	return nil
}

// AtLeast2 .
func AtLeast2(answer interface{}) error {
	answerStr := answer.(string)
	answerInt, err := strconv.Atoi(answerStr)
	if err != nil {
		return fmt.Errorf("Invalid Integer. Got %s", answerStr)
	}
	if answerInt < 2 {
		return fmt.Errorf("Must be at least 2. Got %s", answerStr)
	}
	return nil
}

// IPv4Validator .
func IPv4Validator(ip interface{}) error {
	ipString := ip.(string)
	maybeIP := net.ParseIP(ipString)
	if maybeIP.To4() == nil {
		return fmt.Errorf("Invalid IPv4 address. Got %v", ipString)
	}
	return nil
}

// PortValidator ensures an input is an int between 0 and 65535
// (it could return uint16, but it does not)
func PortValidator(answer interface{}) error {
	answerStr := answer.(string)
	userNum, err := strconv.Atoi(answerStr)
	if err != nil {
		return fmt.Errorf("Invalid port %s\n", answerStr)
	}
	if userNum <= 0 {
		return fmt.Errorf("Port number must be positive. Got %d", userNum)
	}
	if userNum > 65535 {
		return fmt.Errorf("Port number cannot be larger than 65535. Got %d", userNum)
	}
	return nil
}

// IntValidator .
func IntValidator(answer interface{}) error {
	answerStr := answer.(string)
	if _, err := strconv.Atoi(answerStr); err != nil {
		return fmt.Errorf("Invalid Integer. Got %s\n", answerStr)
	}
	return nil
}

// AlwaysValid returns nil as error always
func AlwaysValid(_answer interface{}) error {
	return nil
}

// GroupsRegexValidator .
func GroupsRegexValidator(answer interface{}) error {
	answerStr := answer.(string)
	if !groupsRegex.Match([]byte(answerStr)) {
		return fmt.Errorf("Invalid Groups format. Got %s", answerStr)
	}
	return nil
}

// PositiveIntValidator .
func PositiveIntValidator(answer interface{}) error {
	answerStr := answer.(string)
	num, err := strconv.Atoi(answerStr)
	if err != nil {
		return fmt.Errorf("Invalid Integer. Got %s", answerStr)
	}
	if num <= 0 {
		return fmt.Errorf("Must be positive. Got %s", answerStr)
	}
	return nil
}

// GroupsNumbers returns the numbers of a group
