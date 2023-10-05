package models

type UsersBalance struct {
	UserUUID  string
	Earned    float64
	Current   float64
	Withdrawn float64
}
