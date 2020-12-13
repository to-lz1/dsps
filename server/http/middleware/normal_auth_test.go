package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	. "github.com/saiya/dsps/server/http/testing"
	. "github.com/saiya/dsps/server/jwt/testing"
)

const jwtDir = "../../jwt"
const configRequiresJWT = `
logging:
	debug: false
	category: "*": ERROR

channels:
	-
		regex: 'test-channel'
		jwt:
			iss: [ "https://issuer.example.com/issuer-url" ]
			aud: [ "https://my-service.example.com/" ]
			keys:
				RS256: [ "../../jwt/testdata/RS256-2048bit-public.pem" ]
`

func TestAuthPassing(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(ctx *gin.Context) (Channel, error) {
			return deps.ChannelProvider("test-channel")
		})

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []domain.JwtAud{"https://my-service.example.com/"},
		}))
		ctx.Request = req

		auth(ctx)

		assert.False(t, ctx.IsAborted()) // Should not be aborted
		assert.Equal(t, 200, rec.Code)
	})
}

func TestInvalidChannel(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(ctx *gin.Context) (Channel, error) {
			return deps.ChannelProvider("INVALID-channel") // Invalid channel ID
		})

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		ctx.Request = req

		auth(ctx)

		assert.True(t, ctx.IsAborted())
		AssertRecordedCode(t, rec, http.StatusBadRequest, domain.ErrInvalidChannel)
	})
}

func TestAuthMissingHeader(t *testing.T) {
	WithServerDeps(t, configRequiresJWT+`http: discloseAuthRejectionDetail: true`, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(ctx *gin.Context) (Channel, error) {
			return deps.ChannelProvider("test-channel")
		})

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)

		req, _ := http.NewRequest("GET", "/", nil)
		// No Authorization header
		ctx.Request = req

		auth(ctx)

		assert.True(t, ctx.IsAborted())
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Equal(t, `JWT verification failure: no JWT presented`, GetBodyJSONMap(t, rec)["reason"])
	})
}

func TestAuthRejection(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(ctx *gin.Context) (Channel, error) {
			return deps.ChannelProvider("test-channel")
		})

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		ctx.Request = req

		auth(ctx)

		assert.True(t, ctx.IsAborted())
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Nil(t, GetBodyJSONMap(t, rec)["reason"]) // Should not contain detailed message by default
	})
}

func TestDetailedAuthRejection(t *testing.T) {
	WithServerDeps(t, configRequiresJWT+`http: discloseAuthRejectionDetail: true`, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(ctx *gin.Context) (Channel, error) {
			return deps.ChannelProvider("test-channel")
		})

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		ctx.Request = req

		auth(ctx)

		assert.True(t, ctx.IsAborted())
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Regexp(t, `JWT verification failure.+token is malformed`, GetBodyJSONMap(t, rec)["reason"])
	})
}
