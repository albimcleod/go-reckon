package goreckon

type Contact struct {
	ContactID           string `json:"ContactId"`
	IsEmployee          bool   `json:"IsEmployee"`
	Email               string `json:"Email"`
	FirstNameBranchName string `json:"FirstNameBranchName"`
	SurnameBusinessName string `json:"SurnameBusinessName"`
}
