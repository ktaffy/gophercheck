package testdata

import "fmt"

// BadNestedLoop demonstrates O(n²) complexity - should be detected
func BadNestedLoop(users []User, posts []Post) {
	for i := range users {
		for j := range posts {
			if posts[j].UserID == users[i].ID {
				fmt.Printf("User %s has post %s\n", users[i].Name, posts[j].Title)
			}
		}
	}
}

// BadStringConcat demonstrates inefficient string building - should be detected
func BadStringConcat(items []string) string {
	var result string
	for _, item := range items {
		result += item // This creates new strings each time
	}
	return result
}

// BadSliceSearch demonstrates O(n) search that could be O(1) - should be detected
func BadSliceSearch(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// ComplexFunction has high cyclomatic complexity - should be detected
func ComplexFunction(x, y, z int) string {
	if x > 0 {
		if y > 0 {
			if z > 0 {
				if x > y {
					if y > z {
						if x > 10 {
							if y > 5 {
								return "case1"
							} else {
								return "case2"
							}
						} else {
							return "case3"
						}
					} else {
						return "case4"
					}
				} else {
					return "case5"
				}
			} else {
				return "case6"
			}
		} else {
			return "case7"
		}
	} else {
		return "case8"
	}
}

type User struct {
	ID   int
	Name string
}

type Post struct {
	ID     int
	UserID int
	Title  string
}
