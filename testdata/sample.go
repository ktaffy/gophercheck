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

// BadMemoryAllocationInLoop demonstrates allocation inside loop - should be detected
func BadMemoryAllocationInLoop(data [][]int) [][]int {
	var results [][]int
	for i := 0; i < len(data); i++ {
		temp := make([]int, 10) // Allocates memory each iteration
		for j := 0; j < 10; j++ {
			temp[j] = data[i][j] * 2
		}
		results = append(results, temp)
	}
	return results
}

// BadSliceWithoutCapacity demonstrates slice creation without capacity - should be detected
func BadSliceWithoutCapacity(items []string) []string {
	filtered := make([]string, 0) // Should specify capacity
	for _, item := range items {
		if len(item) > 5 {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// BadMapWithoutSize demonstrates map creation without size hint - should be detected
func BadMapWithoutSize(users []User) map[int]User {
	userMap := make(map[int]User) // Should specify size
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}

// BadAppendInLoop demonstrates append without preallocation - should be detected
func BadAppendInLoop(count int) []int {
	var numbers []int
	for i := 0; i < count; i++ {
		numbers = append(numbers, i*i) // Grows slice each time
	}
	return numbers
}

// GoodMemoryPattern shows the optimized version
func GoodMemoryPattern(data [][]int) [][]int {
	results := make([][]int, 0, len(data)) // Pre-allocate capacity
	temp := make([]int, 10)                // Reuse allocation

	for i := 0; i < len(data); i++ {
		temp = temp[:0] // Reset slice, keep capacity
		for j := 0; j < 10; j++ {
			temp = append(temp, data[i][j]*2)
		}
		// Copy to avoid sharing underlying array
		row := make([]int, len(temp))
		copy(row, temp)
		results = append(results, row)
	}
	return results
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
