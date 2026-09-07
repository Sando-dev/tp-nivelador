package bet

type Bet struct {
	AgencyId int
	Name     string
	LastName string
	Document int
	Birthday string
	Number   int
}

func NewBet(agencyId int, name string, lastName string, document int, birthday string, number int) *Bet {
	return &Bet{
		AgencyId: agencyId,
		Name:     name,
		LastName: lastName,
		Document: document,
		Birthday: birthday,
		Number:   number,
	}
}
