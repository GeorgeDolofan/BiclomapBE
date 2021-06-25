package awscontext

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/savaki/dynastore"
)

type BiclomapAWSContext struct {
	init_err error
	ddb      *dynamodb.DynamoDB
}

const (
	GIN_CONTEXT_KEY_NAME      = "awscontext"
	GIN_CONTEXT_SESSION_STORE = "session-store"
)

var aws_context *BiclomapAWSContext

func init_aws_context() {
	aws_context = new(BiclomapAWSContext)
	session, init_err := session.NewSession()
	if init_err != nil {
		aws_context.init_err = init_err
		log.Println(aws_context.init_err)
		panic(init_err)
	}
	aws_context.ddb = dynamodb.New(session)
	log.Println("awscontext initialized.")
}

func AWSContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		if aws_context == nil {
			init_aws_context()
			// TODO here we should use a build variable in order to sync with
			// the terraform resource creation
			store, err := dynastore.New(
				dynastore.Path("/"),
				dynastore.TableName("biclomap-sessions"),
			)
			if err != nil {
				log.Fatalln(err)
				panic(err)
			}
			c.Set(GIN_CONTEXT_SESSION_STORE, store)
		}
		c.Set(GIN_CONTEXT_KEY_NAME, aws_context)
		c.Next()
	}
}

func GetFromGinContext(c *gin.Context) *BiclomapAWSContext {
	return c.MustGet(GIN_CONTEXT_KEY_NAME).(*BiclomapAWSContext)
}

func GetStore(c *gin.Context) *dynastore.Store {
	return c.MustGet(GIN_CONTEXT_SESSION_STORE).(*dynastore.Store)
}
