package model

type User struct {
	ID          string  `bson:"_id,omitempty" json:"id"`
	Name        string  `bson:"name" json:"name"`
	Email       string  `bson:"email" json:"email"`
	Role        *string `bson:"role,omitempty" json:"role,omitempty"`
	PhoneNumber *string `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
}
