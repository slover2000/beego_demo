package models

// MenuItem represents a menu item which showed in html page
type MenuItem struct {	
	Name     string
	Icon     string
	Children []SubmenuItem
}

// SubmenuItem represents a submen item in the menu item
type SubmenuItem struct {
	ID       int
	Name     string
	Icon     string
	URL      string	
}