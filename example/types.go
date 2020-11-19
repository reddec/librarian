package example

//go:generate librarian -out types_generated.go $GOFILE
type User struct {
	Name string `index:",unique"`
	Role string `index:""`
	SSN  string `index:"bySocialSecurityNum,unique"`
	Year int
}
