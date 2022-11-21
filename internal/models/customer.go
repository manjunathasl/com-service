package models

type Customer struct {
	Email    string
	Text     string
	Schedule string
}

type CustomerSchedules struct {
	List []Customer
}
