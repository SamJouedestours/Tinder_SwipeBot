package logic

import "github.com/you/swipebot/internal/api"

func Decide(c api.Candidate) string {
	for _, it := range c.Interests {
		if it == "climbing" || it == "bouldering" || it == "hiking" {
			return "like"
		}
	}
	if c.Age >= 24 && c.Age <= 35 {
		return "like"
	}
	return "pass"
}
