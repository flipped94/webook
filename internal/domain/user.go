package domain

// 领域对象
type User struct {
	Id        int64
	Email     string
	Phone     string
	Password  string
	Nickname  string
	Birthday  string
	Biography string
	Ctime     int64
}
