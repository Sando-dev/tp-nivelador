package bet


type Bet struct {
	AgencyId int
	Name	 string
	LastName string
	Documentation	int
	Birthday string
	Number   int
}


func NewBet(agencyId int, name string, lastName string, documentation int, birthday string, number int) *Bet {
	return &Bet{
		AgencyId: agencyId,
		Name:     name,
		LastName: lastName,
		Documentation: documentation,
		Birthday: birthday,
		Number:   number,
	}
}