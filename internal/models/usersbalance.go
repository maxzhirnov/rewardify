package models

type UsersBalance struct {
	UserUUID  string
	Earned    float32
	Current   float32
	Withdrawn float32
}
