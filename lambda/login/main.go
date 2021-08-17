package login

import (
	"biclomap-be/lambda/awscontext"
	"bytes"
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
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

type EMAIL_CONFIRM struct {
	TOKEN string `json:"token" binding:"required"`
}

// @Summary Signup new user by their e-mail
// @Description This method starts the new user registration process. It adds
// @Description a new record to the database. The password is stored in hashed form. Then
// @Description it automatically sends an e-mail to the email address given in the parameter
// @Description That e-mail contains a verification token, that's also computed here.
// @Accapt json
// @Produce json
// @Param email_signup body EMAIL_SIGNUP true "Email signup"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /login/email/signup [post]
// @Tags Subscribe
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
	email_confirmation_page := os.Getenv("EMAIL_CONFIRMATION_PAGE")
	if email_confirmation_page == "" {
		log.Println("Please specify the EMAIL_CONFIRMATION_PAGE variable")
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
		biclomap_base_url + "/" + email_confirmation_page + "?token=" + subs_token + "\r\n" +
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

// @Summary Confirms an e-mail
// @Description This method should be called when the user clicks the link
// @Description provided in the e-mail sent by the email signup call. This
// @Description method will effectively activate the user account and, if
// @Description successful, the user will be able to login using the e-mail
// @Description and password provided upon calling the email signup method
// @Param email_confirm body EMAIL_CONFIRM true "email confirmation structure"
// @Accept json
// @Produce json
// @success 200 json map[string]string
// @failure 500 json map[string]string
// @failure 400 json map[string]string
// @failure 404 json map[string]string
// @Router /login/email/confirm [post]
// @Tags Subscribe
// email confirmation will turn an existing signup-XXXXXX UserId into
// email-XXXX. The initial signup-XXXXX record was previously created by the
// EmailSignup method
func EmailConfirm(c *gin.Context) {
	var email_confirm EMAIL_CONFIRM
	err := c.BindJSON(&email_confirm)
	if err != nil {
		log.Println("Failed to bind to input JSON")
		c.JSON(http.StatusNotFound, gin.H{"msg": "Argument error"})
		return
	}
	aws_ctx := awscontext.GetFromGinContext(c)
	delinput := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"UserId": {
				S: aws.String("signup-" + email_confirm.TOKEN),
			},
		},
		ConditionExpression: aws.String("attribute_exists(UserId)"),
		ReturnValues:        aws.String(dynamodb.ReturnValueAllOld),
		TableName:           aws.String("users"),
	}
	old_val, delerr := aws_ctx.Ddb.DeleteItem(delinput)
	if delerr != nil {
		log.Println(delerr.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	old_attributes := old_val.Attributes
	old_attributes["UserId"], _ = dynamodbattribute.Marshal("email-" + email_confirm.TOKEN)
	input := &dynamodb.PutItemInput{
		Item:      old_attributes,
		TableName: aws.String("users"),
	}
	_, dberr := aws_ctx.Ddb.PutItem(input)
	if dberr != nil {
		log.Println("Cannot PutItem into the ddb", dberr)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully confirmed token " + email_confirm.TOKEN})
}

// @Summary Open session for facebook-authenticated user
// @Description This is called after user successfully loged-in with Facebook
// @Accept json
// @Produce json
// @Param fb_login body FB_LOGIN true "login structure"
// @Success 200 {object} LOGIN_INFORMATION
// @Tags Login
// @Router /login/facebook [post]
// TODO this method should be modified in order to also return the JWT token
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

type EMAIL_LOGIN struct {
	EMAIL    string `json:"email" binding:"required"`
	PASSWORD string `json:"password" binding:"required"`
}

type LOGIN_INFORMATION struct {
	Token string `json:"token" binding:"required"`
}

type UserInfo struct {
	UserId   string
	Password []byte `dynamodbav:"password"`
	Salt     []byte `dynamodbav:"salt"`
	Token    string `dynamodbav:"token"`
	Name     string
}

// @Summary Perform email-based login and return JWT token
// @Description Users having succesffuly signed-up and then succesfully having
// @Description confirmed their e-mai will be able to effectively login into
// @Description the application using this method call. The system will create
// @Description a JWT token. Any subsequent API call will need to provide this
// @Description token in the request body.
// @Tags Login
// @Accept json
// @Produce json
// @Param email_login body EMAIL_LOGIN true "Email and password structure"
// @success 200 {object} LOGIN_INFORMATION
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /login/email [post]
func Email(c *gin.Context) {
	var email_login EMAIL_LOGIN
	bind_err := c.BindJSON(&email_login)
	if bind_err != nil {
		log.Println("Failed to bind to input JSON")
		c.JSON(http.StatusNotFound, gin.H{"msg": "Argument error"})
		return
	}
	if !strings.Contains(email_login.EMAIL, "@") {
		log.Println("EMail looks to be incorrect", email_login.EMAIL)
		c.JSON(http.StatusNotFound, gin.H{"msg": "Argument error"})
		return
	}
	if len(email_login.PASSWORD) == 0 {
		log.Println("Password cannot be empty")
		c.JSON(http.StatusNotFound, gin.H{"msg": "Password cannot be empty"})
		return
	}
	// lookup user in the database
	aws_ctx := awscontext.GetFromGinContext(c)
	input := &dynamodb.QueryInput{
		IndexName:                aws.String("EmailIndex"),
		KeyConditionExpression:   aws.String("email = :v_email"),
		ExpressionAttributeNames: map[string]*string{"#T": aws.String("token"), "#N": aws.String("Name")},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v_email": {S: aws.String(email_login.EMAIL)},
		},
		ProjectionExpression: aws.String("UserId, password, salt, #T, #N"),
		TableName:            aws.String("users"),
	}
	res, err := aws_ctx.Ddb.Query(input)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Println("Query returns: ", res)
	if *res.Count != 1 {
		log.Println("OOPS, we found more than 1 user with these credentials", email_login.EMAIL)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Unexpected multiple accounts for this user"})
		return
	}
	userInfo := UserInfo{}
	read_err := dynamodbattribute.UnmarshalMap(res.Items[0], &userInfo)
	if read_err != nil {
		log.Println("Cannot UnmarshalMap: ", read_err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Unexpected error while accessing user account"})
		return
	}
	// check passwords do match
	log.Println("UserInfo: ", userInfo)
	hash_password := pbkdf2.Key([]byte(email_login.PASSWORD), userInfo.Salt, 4096, sha256.Size, sha256.New)
	if !bytes.Equal(hash_password, userInfo.Password) {
		log.Println("Passwords do not match")
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Incorrect credentials"})
		return
	}
	// prepare the JWT token
	jwtWrapper := JwtWrapper{
		SecretKey:       "xxxx", // TODO use a key extracted from the ENV here
		Issuer:          "biclomap",
		ExpirationHours: 24,
	}
	jwt_token, jwt_err := jwtWrapper.GenerateToken(email_login.EMAIL, userInfo.UserId)
	if jwt_err != nil {
		log.Println("Cannot create JWT token", jwt_err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "JWT Error"})
		return
	}

	login_information := LOGIN_INFORMATION{
		Token: jwt_token,
	}
	c.JSON(http.StatusOK, login_information)
}
