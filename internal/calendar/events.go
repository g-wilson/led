package calendar

var events = eventList{
	{"AUS", "2026-03-08T04:00:00Z", f1Image}, // Australian Grand Prix (Melbourne, AEDT 15:00)
	{"CHN", "2026-03-15T07:00:00Z", f1Image}, // Chinese Grand Prix (Shanghai, CST 15:00)
	{"JPN", "2026-03-29T05:00:00Z", f1Image}, // Japanese Grand Prix (Suzuka, JST 14:00)
	{"BHR", "2026-04-12T15:00:00Z", f1Image}, // Bahrain Grand Prix (Sakhir, AST 18:00)
	{"SAU", "2026-04-19T17:00:00Z", f1Image}, // Saudi Arabian Grand Prix (Jeddah, AST 20:00)
	{"MIA", "2026-05-03T20:00:00Z", f1Image}, // Miami Grand Prix (Miami, EDT 16:00)
	{"CAN", "2026-05-24T20:00:00Z", f1Image}, // Canadian Grand Prix (Montreal, EDT 16:00)
	{"MON", "2026-06-07T13:00:00Z", f1Image}, // Monaco Grand Prix (CEST 15:00)
	{"CAT", "2026-06-14T13:00:00Z", f1Image}, // Barcelona-Catalunya Grand Prix (CEST 15:00)
	{"AUT", "2026-06-28T13:00:00Z", f1Image}, // Austrian Grand Prix (Spielberg, CEST 15:00)
	{"GBR", "2026-07-05T14:00:00Z", f1Image}, // British Grand Prix (Silverstone, BST 15:00)
	{"BEL", "2026-07-19T13:00:00Z", f1Image}, // Belgian Grand Prix (Spa, CEST 15:00)
	{"HUN", "2026-07-26T13:00:00Z", f1Image}, // Hungarian Grand Prix (Budapest, CEST 15:00)
	{"NED", "2026-08-23T13:00:00Z", f1Image}, // Dutch Grand Prix (Zandvoort, CEST 15:00)
	{"ITA", "2026-09-06T13:00:00Z", f1Image}, // Italian Grand Prix (Monza, CEST 15:00)
	{"ESP", "2026-09-13T13:00:00Z", f1Image}, // Spanish Grand Prix (Madrid, CEST 15:00)
	{"AZE", "2026-09-26T11:00:00Z", f1Image}, // Azerbaijan Grand Prix (Baku, AZT 15:00, Saturday race)
	{"SGP", "2026-10-11T12:00:00Z", f1Image}, // Singapore Grand Prix (SGT 20:00)
	{"USA", "2026-10-25T20:00:00Z", f1Image}, // United States Grand Prix (Austin, CDT 15:00)
	{"MEX", "2026-11-01T20:00:00Z", f1Image}, // Mexican Grand Prix (Mexico City, CST 14:00)
	{"BRA", "2026-11-08T17:00:00Z", f1Image}, // Brazilian Grand Prix (São Paulo, BRT 14:00)
	{"LAS", "2026-11-22T04:00:00Z", f1Image}, // Las Vegas Grand Prix (PST 20:00 Sat, UTC Sun)
	{"QAT", "2026-11-29T16:00:00Z", f1Image}, // Qatar Grand Prix (Lusail, AST 19:00)
	{"UAE", "2026-12-06T13:00:00Z", f1Image}, // Abu Dhabi Grand Prix (Yas Marina, GST 17:00)

	{"Eurovision", "2026-05-16T19:00:00Z", nil}, // Eurovision Song Contest

	{"XMAS", "2026-12-25T00:00:00.000Z", xmasImage},
	{"2027", "2027-01-01T00:00:00.000Z", xmasImage},
}
