package imgix

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Idiomatic testing.
func testClient() Builder {
	return NewBuilder("prod.imgix.net")
}

func testClientWithToken() Builder {
	return NewBuilderWithToken("my-social-network.imgix.net", "FOO123bar")
}

func TestBasicClientPath(t *testing.T) {
	c := testClient()
	assert.Equal(t, "https://prod.imgix.net/1/users.jpg", c.CreateURLFromPath("/1/users.jpg"))
}

func TestClientPathWithParams(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"200"}, "h": []string{"400"}}
	assert.Equal(t, "https://prod.imgix.net/1/users.jpg?h=400&w=200", c.CreateURL("/1/users.jpg", params))
}

func TestClientScheme(t *testing.T) {
	c := testClient()
	c.useHTTPS = false
	assert.Equal(t, "http", c.Scheme())
	c.useHTTPS = true
	assert.Equal(t, "https", c.Scheme())
}

func TestClientPath(t *testing.T) {
	c := testClient()
	u := c.CreateURLFromPath("/jax.jpg")
	assert.Equal(t, "https://prod.imgix.net/jax.jpg", u)
}

func TestClientPathWithOneParam(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"400"}}
	u := c.CreateURL("/jax.jpg", params)
	assert.Equal(t, "https://prod.imgix.net/jax.jpg?w=400", u)
}

func TestClientPathWithTwoParams(t *testing.T) {
	c := testClient()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	u := c.CreateURL("/jax.jpg", params)
	assert.Equal(t, "https://prod.imgix.net/jax.jpg?h=300&w=400", u)
}

func TestClientPathWithParamsEncodesParamKeys(t *testing.T) {
	c := testClient()
	params := url.Values{"hello world": []string{"interesting"}}
	u := c.CreateURL("/demo.png", params)
	assert.Equal(t, "https://prod.imgix.net/demo.png?hello%%20world=interesting", u)
}

func TestClientPathWithParamsEncodesParamValues(t *testing.T) {
	c := testClient()
	params := url.Values{"hello_world": []string{"/foo\"> <script>alert(\"hacked\")</script><"}}
	u := c.CreateURL("/demo.png", params)
	assert.Equal(t, "https://prod.imgix.net/demo.png?hello_world=%2Ffoo%22%3E%%20%3Cscript%3Ealert%28%22hacked%22%29%3C%2Fscript%3E%3C", u)
}

func TestClientPathWithParamsEncodesBase64ParamVariants(t *testing.T) {
	c := testClient()
	params := url.Values{"txt64": []string{"I cannøt belîév∑ it wors! 😱"}}
	u := c.CreateURL("~text", params)
	assert.Equal(t, "https://prod.imgix.net/~text?txt64=SSBjYW5uw7h0IGJlbMOuw6l24oiRIGl0IHdvcu-jv3MhIPCfmLE", u)
}

func TestClientPathWithSignature(t *testing.T) {
	c := testClientWithToken()
	u := c.CreateSignedURLFromPath("/users/1.png")
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", u)
}

func TestClientPathWithSignatureAndParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?h=300&w=400&s=9de08f728192f92b7132176a0a17ef08", c.CreateSignedURL("/users/1.png", params))
}

func TestClientPathWithSignatureAndEmptyParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{}
	assert.Equal(t, "https://my-social-network.imgix.net/users/1.png?s=6797c24146142d5b40bde3141fd3600c", c.CreateSignedURL("/users/1.png", params))
}

func TestClientFullyQualifiedUrlPath(t *testing.T) {
	c := testClientWithToken()
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?s=493a52f008c91416351f8b33d4883135", c.CreateSignedURLFromPath("http://avatars.com/john-smith.png"))
}

func TestClientFullyQualifiedUrlPathWithParams(t *testing.T) {
	c := testClientWithToken()
	params := url.Values{"w": []string{"400"}, "h": []string{"300"}}
	assert.Equal(t, "https://my-social-network.imgix.net/http%3A%2F%2Favatars.com%2Fjohn-smith.png?h=300&w=400&s=a58d87a5dfa5f7478e06715571e96f78", c.CreateSignedURL("http://avatars.com/john-smith.png", params))
}

func TestClientHostsCountValidation(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			assert.True(t, ok)
			assert.EqualError(t, e, "hosts must be provided")
		}
	}()

	c := testClient()
	c.domain = string("")
	c.Domain()
}

func TestCreateSrcSetWidths(t *testing.T) {
	c := testClient()
	actual := c.CreateSrcSetFromWidths("image.jpg", url.Values{}, []int{100, 200, 300, 400})
	expected := "https://prod.imgix.net/image.jpg?w=100 100w\n" +
		"https://prod.imgix.net/image.jpg?w=200 200w\n" +
		"https://prod.imgix.net/image.jpg?w=300 300w\n" +
		"https://prod.imgix.net/image.jpg?w=400 400w"
	assert.Equal(t, expected, actual)
}
