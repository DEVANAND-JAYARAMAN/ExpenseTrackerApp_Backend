package main

// CategoryMap represents predefined categories with fixed IDs
var CategoryMap = map[int]string{
	1:  "Food",
	2:  "Transportation", 
	3:  "Entertainment",
	4:  "Shopping",
	5:  "Bills",
	6:  "Fuel",
	7:  "School_Fees",
	8:  "Medical",
	9:  "Rent",
	10: "Utilities",
	11: "Insurance",
	12: "Clothing",
	13: "Travel",
	14: "Gym",
	15: "Books",
	16: "Electronics",
	17: "Home_Maintenance",
	18: "Pet_Care",
	19: "Gifts",
	20: "Charity",
}

// GetCategoryName returns category name by ID
func GetCategoryName(categoryID int) (string, bool) {
	name, exists := CategoryMap[categoryID]
	return name, exists
}

// GetAllCategories returns all available categories
func GetAllCategories() []CategoryItem {
	var categories []CategoryItem
	for id, name := range CategoryMap {
		categories = append(categories, CategoryItem{
			ID:   id,
			Name: name,
		})
	}
	return categories
}

// CategoryItem represents a category for frontend
type CategoryItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}