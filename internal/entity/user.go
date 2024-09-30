package entity

// User represents a user in the system
type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}
