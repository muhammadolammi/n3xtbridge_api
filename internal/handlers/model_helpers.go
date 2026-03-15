package handlers

import "github.com/muhammadolammi/n3xtbridge_api/internal/database"

func dbUserToUser(dbUser database.User) User {
	return User{
		ID:          dbUser.ID,
		FirstName:   dbUser.FirstName,
		LastName:    dbUser.LastName,
		Email:       dbUser.Email,
		PhoneNumber: dbUser.PhoneNumber,
		Address:     dbUser.Address,
		Role:        dbUser.Role,
	}

}
