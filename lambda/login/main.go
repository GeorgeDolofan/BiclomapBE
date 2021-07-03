package login

import (
	"biclomap-be/lambda/awscontext"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

type FB_LOGIN struct {
	NAME string `json:"name" binding:"required"`
	ID   string `json:"id" binding:"required"`
}

type Facebook_User struct {
	UserId string
	Name   string
}

type EMAIL_SIGNUP struct {
	EMAIL    string `json:"email" binding:"required"`
	PASSWORD string `json:"password" binding:"required"`
}

func EmailSignup(c *gin.Context) {
	var email_signup EMAIL_SIGNUP
	err := c.BindJSON(&email_signup)
	if err != nil {
		log.Println("Failed to bind to input JSON")
		c.JSON(http.StatusNotFound, gin.H{"msg": "Argument error"})
		return
	}
	if !strings.Contains(email_signup.EMAIL, "@") {
		log.Println("EMail looks to be incorrect", email_signup.EMAIL)
		c.JSON(http.StatusNotFound, gin.H{"msg": "Argument error"})
		return
	}
	if len(email_signup.PASSWORD) == 0 {
		log.Println("Password cannot be empty")
		c.JSON(http.StatusNotFound, gin.H{"msg": "Password cannot be empty"})
		return
	}

	secretName := "biclomap_smtp_password"
	region := "eu-central-1"

	//Create a Secrets Manager client
	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	secret_input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(secret_input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				log.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				log.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				log.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				log.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				log.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		c.AbortWithError(500, err)
		return
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secret_value string
	if result.SecretString != nil {
		secret_value = *result.SecretString
	} else {
		secret_valueBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(secret_valueBytes, result.SecretBinary)
		if err != nil {
			log.Println("Base64 Decode Error:", err)
			c.AbortWithError(500, err)
			return
		}
		secret_value = string(secret_valueBytes[:len])
	}
	var key_val map[string]interface{}
	json.Unmarshal([]byte(secret_value), &key_val)
	smtp_password := key_val["biclomap_smtp_password"].(string)

	smtp_server := os.Getenv("SMTP_SERVER")
	if smtp_server == "" {
		log.Println("Please specify SMTP_SERVER variable")
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Server incorrectly configured, please contact the administrator"})
		return
	}
	biclomap_base_url := os.Getenv("BICLOMAP_BASE_URL")
	if biclomap_base_url == "" {
		log.Println("Please specify the BICLOMAP_BASE_URL variable")
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Server incorrectly configured, please contact the administrator"})
		return
	}

	// ok, everything looks to be correctly configured so far, so let's proceed
	// to store this new subscription attempt and then send the e-mail
	subs_token := uuid.New().String()

	salt := make([]byte, 32)
	_, readerr := io.ReadFull(rand.Reader, salt)
	if readerr != nil {
		log.Println("Please ")
	}
	hash_password := pbkdf2.Key([]byte(email_signup.PASSWORD), salt, 4096, sha256.Size, sha256.New)

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"UserId": {
				S: aws.String("signup-" + subs_token),
			},
			"email": {
				S: aws.String(email_signup.EMAIL),
			},
			"salt": {
				B: salt,
			},
			"password": {
				B: hash_password,
			},
			"token": {
				S: aws.String(subs_token),
			},
			"timestamp": {
				S: aws.String(time.Now().UTC().Format(time.RFC3339)),
			},
		},
		TableName: aws.String("users"),
	}
	aws_ctx := awscontext.GetFromGinContext(c)
	_, dberr := aws_ctx.Ddb.PutItem(input)
	if dberr != nil {
		log.Println("Cannot PutItem into the ddb", dberr)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Database error"})
		return
	}

	auth := smtp.PlainAuth("", "notifications@biclomap.com", smtp_password, smtp_server)
	to := []string{email_signup.EMAIL}
	msg := []byte("To:" + email_signup.EMAIL + "\r\n" +
		"Subject: Biclomap Subscription Confirmation\r\n" +
		"\r\n" +
		"Hi!\r\n" +
		"\r\n" +
		"Somobody - probably you - signed you-up to Biclomap.\r\n" +
		"If that's correct, then please click the link below\r\n" +
		"to confirm your e-mail address:\r\n" +
		biclomap_base_url + "/email-confirmation?token=" + subs_token + "\r\n" +
		"\r\n\r\n" +
		"Cheers,\r\n" +
		"The Biclomap Team")
	mail_err := smtp.SendMail(smtp_server+":587", auth, "notifications@biclomap.com", to, msg)
	if mail_err != nil {
		log.Println(mail_err)
		c.AbortWithError(500, mail_err)
	} else {
		log.Println("Successfully sent message")
		c.JSON(http.StatusOK, gin.H{"msg": "confirmation mail sent"})
	}
}

// @summary Used to authenticate a Facebook user
// @Accept json
// @Produce json
// @Param fb_login body FB_LOGIN true "login structure"
func Facebook(c *gin.Context) {
	var fb_login FB_LOGIN
	err := c.BindJSON(&fb_login)
	if err == nil {
		aws_ctx := awscontext.GetFromGinContext(c)
		input := &dynamodb.GetItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"UserId": {
					S: aws.String("fb-" + fb_login.ID),
				},
			},
			TableName: aws.String("users"),
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
						S: aws.String("fb-" + fb_login.ID),
					},
					"Name": {
						S: aws.String(fb_login.NAME),
					},
				},
				TableName: aws.String("users"),
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
