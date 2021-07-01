package login

import (
	"biclomap-be/lambda/awscontext"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
TODO create authorization token
*/

// @summary Used to authenticate a Facebook user
// @Accept json
// @Produce text/plain
// @Param fb_login body FB_LOGIN true "login structure"
func Facebook(c *gin.Context) {
	var fb_login FB_LOGIN
	err := c.BindJSON(&fb_login)
	if err == nil {
		aws_ctx := awscontext.GetFromGinContext(c)
		input := &dynamodb.GetItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"UserId": {
					S: aws.String(fb_login.ID),
				},
			},
			TableName: aws.String("facebook-users"),
		}
		res, err := aws_ctx.Ddb.GetItem(input)
		if err != nil {
			log.Println(err.Error())
			c.AbortWithStatus(500)
			return
		}
		var msg struct {
			Session_kind string `json:"session-kind"`
			Session_id   string `json:"session-id"`
			New_user     string `json:"new_user"`
		}
		msg.Session_kind = "facebook"
		msg.Session_id = fb_login.ID
		if len(res.Item) == 0 {
			input := &dynamodb.PutItemInput{
				Item: map[string]*dynamodb.AttributeValue{
					"UserId": {
						S: aws.String(fb_login.ID),
					},
					"Name": {
						S: aws.String(fb_login.NAME),
					},
				},
				TableName: aws.String("facebook-users"),
			}
			_, err := aws_ctx.Ddb.PutItem(input)
			if err != nil {
				log.Println(err.Error)
				c.AbortWithStatus(500)
				return
			}
			log.Println("New facebook user")
			msg.New_user = "1"
		} else {
			log.Println("Existing facebook user")
			msg.New_user = "0"
		}

		store := awscontext.GetStore(c)
		session, err := store.Get(c.Request, "session.id")
		if err != nil {
			log.Println("ERROR: Cannot create session")
			panic(err)
		}
		session.Values["session-info"] = msg
		session.Save(c.Request, c.Writer)
		c.JSON(http.StatusOK, msg)
	}
}

func Email(c *gin.Context) {
	c.String(404, "Not yet implemented")
}
