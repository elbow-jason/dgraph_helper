package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var dashAndCommaRegex = regexp.MustCompile("(,|-)")

// https://stackoverflow.com/questions/6878590/the-maximum-value-for-an-int-type-in-go
const maxUint = ^uint(0)
const maxInt = int(maxUint >> 1)
const minInt = -maxInt - 1

// Returns the numbers that appeared in the groups.
// Not the range of groups. Just the numbers.
// the intent is to check the max number is less
// than or equal to the number of groups - 1 in the config
func splitGroups(groups string) []int {
	nums := []int{}
	parts := dashAndCommaRegex.Split(groups, -1)
	for _, num := range parts {
		integer, err := strconv.Atoi(num)
		if err != nil {
			// if this ever panics it is because the
			// prompts.GroupsRegexValidator failed to screen
			// improperly formatted groups.
			panic(err)
		}
		nums = append(nums, integer)
	}
	return nums
}

func maxIntOfSlice(nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, fmt.Errorf("Could not get max of empty slice")
	}
	max := minInt
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max, nil
}

func int2string(num int) string {
	return strconv.Itoa(num)
}

func float2string(num float64) string {
	return fmt.Sprintf("%.2f", num)
}

func bool2string(b bool) string {
	return strconv.FormatBool(b)
}

func runCommand(cmds ...string) error {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
