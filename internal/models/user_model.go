package models

import "time"

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginData struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type UserSession struct {
	UserId     string     `json:"user_id"`
	Token      string     `json:"token"`
	LoginTime  *time.Time `json:"login_time"`
	LogoutTime *time.Time `json:"logout_time"`
}

type User struct {
	Id           string       `json:"id"`
	Username     string       `json:"username" binding:"required"`
	Password     string       `json:"password" binding:"required"`
	FullName     string       `json:"full_name" binding:"required"`
	Status       int          `json:"status" binding:"required"`
	Level        int          `json:"level"`
	CustomerData CustomerData `gorm:"foreignKey:user_id;reference:id" json:"customer_data" `
	CreatedAt    *time.Time   `json:"created_at"`
	UpdatedAt    *time.Time   `json:"updated_at"`
}

type CustomerData struct {
	UserId           string     `json:"user_id"`
	Dob              string     `json:"dob"`
	PhoneNumber      string     `json:"phone_number"`
	Gender           string     `json:"gender"`
	MaritalStatus    string     `json:"marital_status"`
	Address          string     `json:"address"`
	DistrictAddress  string     `json:"district_address"`
	CityAddress      string     `json:"city_address"`
	ProvinceAddress  string     `json:"province_address"`
	PostalCode       int        `json:"postal_code"`
	LatitudeAddress  string     `json:"latitude_address"`
	LongitudeAddress string     `json:"longitude_address"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}
