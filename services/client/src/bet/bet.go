package bet


type Bet struct {
	AgencyId int
	Name	 string
	LastName string
	DNI	int
	Birthday string
	Amount   int
}


func NewBet(agencyId int, name string, lastName string, dni int, birthday string, amount int) *Bet {
	return &Bet{
		AgencyId: agencyId,
		Name:     name,
		LastName: lastName,
		DNI:      dni,
		Birthday: birthday,
		Amount:   amount,
	}
}