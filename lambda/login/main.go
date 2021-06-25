package login

import (
	"biclomap-be/lambda/awscontext"
	"log"

	"github.com/gin-gonic/gin"
)

type FB_LOGIN struct {
	NAME string `json:"name" binding:"required"`
	ID   string `json:"id" binding:"required"`
}

type Facebook_User struct {
	UserId string
	Name   string
}

/*
TODO REST API should never send back cookies so this should be changed
  FB_LOGIN session is used to lookup the table facebook-users
	  if not found, then
			create the user
			create the identity record
			return 201 along with the authorization token
		if found then return 200 with the authorization token
*/

// @summary Used to authenticate a Facebook user
// @Accept json
// @Produce text/plain
// @Param fb_login body FB_LOGIN true "login structure"
func Facebook(c *gin.Context) {
	var fb_login FB_LOGIN
	err := c.BindJSON(&fb_login)
	if err == nil {
		store := awscontext.GetStore(c)
		session, err := store.Get(c.Request, "session.id")
		if err != nil {
			log.Println("ERROR: Cannot create session")
			panic(err)
		}
		eventually_authenticated := session.Values["authenticated"]
		if eventually_authenticated != nil && eventually_authenticated != false {
			c.String(302, "Already Authenticated")
			return
		}
		session.Values["authenticated"] = true
		session.Save(c.Request, c.Writer)
		c.String(201, "Session created")
	} else {
		log.Printf("ERROR: %s", err.Error())
		c.String(400, err.Error())
	}
}

func Email(c *gin.Context) {
	c.String(404, "Not yet implemented")
}
