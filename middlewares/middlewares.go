// Copyright (c) 2019 JasaCloud.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middlewares

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jasacloud/go-libraries/apiv3"
	"github.com/jasacloud/go-libraries/client"
	"github.com/jasacloud/go-libraries/db"
	"github.com/jasacloud/go-libraries/server"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

// DbConnect function
// Connect middleware clones the database session for each request and
// makes the `db` object available for each handler
func DbConnect(c *gin.Context) {
	s := db.Properties.Sess.Clone()

	defer s.Close()

	c.Set("db", s.DB(db.Properties.Conn.Name))
	c.Next()
}

// ErrorHandler is a middleware to handle errors encountered during requests
func ErrorHandler(c *gin.Context) {
	c.Next()
	// TODO: Handle it in a better way
	if len(c.Errors) > 0 {
		errorCode := "400"
		if e, exists := c.Get("error"); exists {
			response := gin.H{
				"returnval": false,
				"error":     e,
			}
			c.AbortWithStatusJSON(200, response)
		} else {
			errorVal := gin.H{
				"code":    errorCode,
				"type":    "request",
				"message": c.Errors[0].Error(),
			}
			if len(c.Errors) > 1 {
				errorVal["message_details"] = c.Errors[1].Error()
			}
			if errorCode, exists := c.Get("error_code"); exists {
				errorVal["code"] = errorCode
			}
			response := gin.H{
				"returnval": false,
				"error":     errorVal,
			}
			c.AbortWithStatusJSON(200, response)
		}

		return
	}
}

// BindJSON function
func BindJSON(c *gin.Context, obj interface{}) error {
	if err := binding.JSON.Bind(c.Request, obj); err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypeBind)
		return err
	}
	return nil
}

// AuthenticatedHeader function
func AuthenticatedHeader(c *gin.Context) bool {
	if v := c.GetHeader("X-Token-Audience"); v != "" {
		claim := apiv3.Claims{
			Aud:      v,
			ClientId: v,
		}
		if v := c.GetHeader("X-Token-Credential"); v != "" {
			claim.Cre = v
			claim.CredentialId = v
		}
		if v := c.GetHeader("X-Token-Subject"); v != "" {
			claim.UserId = v
			claim.Sub = v
		}
		if v := c.GetHeader("X-Token-Issuer"); v != "" {
			claim.Iss = v
		}
		if v := c.GetHeader("X-Token-Issued-At"); v != "" {
			claim.Iat, _ = strconv.Atoi(v)
		}
		if v := c.GetHeader("X-Token-Expired-At"); v != "" {
			claim.Exp, _ = strconv.Atoi(v)
			claim.Expired, _ = strconv.Atoi(v)
		}

		c.Set("claims", claim)
		return true
	}
	return false
}

// AuthServer function
func AuthServer(defaultOpt client.Http) gin.HandlerFunc {
	return func(c *gin.Context) {
		if AuthenticatedHeader(c) {
			c.Next()
			return
		}

		opt := client.LoadHttp(defaultOpt.HttpResource.Url)
		//when authorization source header defined from request :
		if c.GetHeader("Authorization-Source") != "" {
			opt = client.LoadHttpResource(c.GetHeader("Authorization-Source"))
		}
		opt.SetRequest("GET", "", nil)
		if opt.Err != nil {
			response := gin.H{
				"returnval": false,
				"error": gin.H{
					"code":    "501",
					"type":    "AuthServer",
					"message": "Bad Server",
				},
			}
			c.JSON(200, response)
			c.Abort()
			return
		}
		if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer") {
			opt.SetHeader("Authorization", c.GetHeader("Authorization"))
		} else {
			opt.SetHeader("Authorization", "Bearer "+c.GetHeader("Authorization"))
		}

		if c.GetHeader("Authorization-Source") != "" {
			opt.SetHeader("Authorization-Source", c.GetHeader("Authorization-Source"))
		}
		resp, err := opt.Start()
		if err != nil {
			response := gin.H{
				"returnval": false,
				"error": gin.H{
					"code":    "501",
					"type":    "AuthServer",
					"message": "Bad Gateway",
				},
			}
			c.JSON(200, response)
			c.Abort()
			return
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		ProcessAuthResponse(c, resp.Body)
	}
}

// Auth  function
func Auth(opt server.JwtOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		//when authorization source header defined from request :
		if c.GetHeader("Authorization-Source") != "" {
			opt = server.GetJwtOption(c.GetHeader("Authorization-Source"))
		} else {
			opt = server.GetJwtOption(server.DefaultResourceName)
		}
		token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			var b interface{}
			if (strings.HasPrefix(token.Method.Alg(), "RS") || strings.HasPrefix(token.Method.Alg(), "PS")) && opt.PublicKey != "" {
				if server.PublicKey[opt.PublicKey] == nil {
					server.PublicKey[opt.PublicKey], _ = ioutil.ReadFile(opt.PublicKey)
					b, _ = jwt.ParseRSAPublicKeyFromPEM(server.PublicKey[opt.PublicKey].([]byte))
				} else {
					b, _ = jwt.ParseRSAPublicKeyFromPEM(server.PublicKey[opt.PublicKey].([]byte))
				}
			} else {
				b = ([]byte(opt.Secret))
			}
			return b, nil
		})

		if err != nil {
			response := gin.H{
				"returnval": false,
				"error": gin.H{
					"code":    "401",
					"type":    "Authentication",
					"message": "Request Unauthorized",
				},
			}
			c.JSON(200, response)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("claims", claims)

			//Add jwt claim to c.Params, to usage add the c.Param("jwt_"+claimName) :
			for i, v := range claims {
				c.Set("jwt_"+i, v)
				switch v := v.(type) {
				case float64:
					c.Params = append(c.Params, gin.Param{"jwt_" + i, strconv.FormatFloat(v, 'f', 0, 64)})
				case int:
					c.Params = append(c.Params, gin.Param{"jwt_" + i, strconv.Itoa(v)})
				case string:
					c.Params = append(c.Params, gin.Param{"jwt_" + i, v})
				default:
					fmt.Printf("Unknow Type of Params: %T\n", v)
				}
			}
		} else {
			response := gin.H{
				"returnval": false,
				"error": gin.H{
					"code":    "401",
					"type":    "Authentication",
					"message": "Request Unauthorized",
				},
			}
			c.JSON(200, response)
			c.Abort()
			return
		}

		//To get on handler :
		//calims:=c.MustGet("claims").(jwt.MapClaims)
		c.Set("token", token)
		c.Next()
	}
}

// CreateToken function
func CreateToken(opt server.JwtOption, claims jwt.MapClaims) string {
	allgoritm := opt.Algorithm
	var key interface{}
	if (strings.HasPrefix(allgoritm, "RS") || strings.HasPrefix(allgoritm, "PS")) && opt.PrivateKey != "" {
		if server.PrivateKey[opt.PrivateKey] == nil {
			server.PrivateKey[opt.PrivateKey], _ = ioutil.ReadFile(opt.PrivateKey)
			key, _ = jwt.ParseRSAPrivateKeyFromPEM(server.PrivateKey[opt.PrivateKey].([]byte))
		} else {
			key, _ = jwt.ParseRSAPrivateKeyFromPEM(server.PrivateKey[opt.PrivateKey].([]byte))
		}
	} else {
		key = []byte(opt.Secret)
	}
	token := jwt.New(jwt.GetSigningMethod(allgoritm))
	claims["iat"] = time.Now().Unix()
	var exp int64
	if opt.ExpirationSec > 0 {
		exp = time.Now().Add(time.Second * time.Duration(opt.ExpirationSec)).Unix()
	} else {
		exp = time.Now().Add(time.Hour * 1).Unix()
	}
	claims["exp"] = exp
	token.Claims = claims

	tokenString, err := token.SignedString(key)
	if err != nil {
		return ""
	}

	return tokenString
}

// GetToken function
func GetToken(c *gin.Context) {
	claims := jwt.MapClaims{}
	if c.Param("userId") != "" {
		claims["userid"] = c.Param("userId")
	} else {
		claims["userid"] = "110148036004494553296"
	}

	token := CreateToken(server.GetJwtOption(c.Param("source")), claims)

	response := gin.H{
		"returnval": true,
		"kind":      "get#token",
		"result":    token,
	}

	//start response sent :
	c.Writer.Header().Set("x-response-sent", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	c.JSON(200, response)
}

// ProcessAuthResponse function
func ProcessAuthResponse(c *gin.Context, body io.Reader) {
	data, err := server.GinUnmarshal(body)
	if err != nil {
		response := gin.H{
			"returnval": false,
			"error": gin.H{
				"code":    "501",
				"type":    "AuthServer",
				"message": "Bad Response Level 1",
			},
		}
		c.JSON(200, response)
		c.Abort()
		return
	}
	if data["returnval"] == true {
		values, err := server.GinReUnmarshal(data["values"])
		if err != nil {
			response := gin.H{
				"returnval": false,
				"error": gin.H{
					"code":    "501",
					"type":    "AuthServer",
					"message": "Bad Response Level 1",
				},
			}
			c.JSON(200, response)
			c.Abort()
			return
		}
		if values["credential_id"] != nil {
			c.Set("credential_id", values["credential_id"])
			if credentialId, ok := values["credential_id"].(string); ok {
				c.Params = append(c.Params, gin.Param{"credential_id", credentialId})
			} else {
				log.Println("credential_id not found in claims")
			}
		}
		if values["user_id"] != nil {
			c.Set("user_id", values["user_id"])
			if userId, ok := values["user_id"].(string); ok {
				c.Params = append(c.Params, gin.Param{"jwt_sub", userId})
				c.Params = append(c.Params, gin.Param{"user_id", userId})
			} else {
				log.Println("user_id not found in claims")
			}
		}
		if values["claims"] != nil {
			c.Set("claims", values["claims"])
		}
		c.Next()
	} else {
		response := gin.H{
			"returnval": false,
			"error": gin.H{
				"code":    "401",
				"type":    "Request",
				"message": "Request Unauthorized",
			},
		}
		c.JSON(200, response)
		c.Abort()
		return
	}
}
