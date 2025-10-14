package models

type BarData struct {
	NodeDetails
	OSMNode       string
	FormattedAddr string
	Street        string
	LinkedBar     bool
	LiveMusic     bool
	Food          bool
}

type AddrIDs struct {
	Fylke   int
	Kommune int
	Sted    int
}
