package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type onepasswordTestCase struct {
	Prompt      bool
	EnvironFunc func() []string
}

//nolint:dupl,forcetypeassert
func TestOnepasswordTemplateFuncV1(t *testing.T) {
	for i, tc := range onepasswordTestCases(t) {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &Config{
				ConfigFile: ConfigFile{
					Onepassword: onepasswordConfig{
						Command:     "op",
						Prompt:      tc.Prompt,
						environFunc: tc.EnvironFunc,
					},
				},
				baseSystem: &testSystem{
					System:     chezmoi.NullSystem{},
					outputFunc: onepasswordV1OutputFunc,
				},
			}

			require.NotPanics(t, func() {
				actual := c.onepasswordTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["uuid"], "wxcplh5udshnonkzg2n4qx262y")
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordDetailsFieldsTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["password"].(map[string]interface{})["value"], "L8rm1JXJIE1b8YUDWq7h")
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordDocumentTemplateFunc("ExampleDocument")
				assert.Equal(t, "ExampleDocumentContents", actual)
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordItemFieldsTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["exampleLabel"].(map[string]interface{})["v"], "exampleValue")
			})
		})
	}
}

//nolint:dupl,forcetypeassert
func TestOnepasswordTemplateFuncV2(t *testing.T) {
	for i, tc := range onepasswordTestCases(t) {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &Config{
				ConfigFile: ConfigFile{
					Onepassword: onepasswordConfig{
						Command:     "op",
						Prompt:      tc.Prompt,
						environFunc: tc.EnvironFunc,
					},
				},
				baseSystem: &testSystem{
					System:     chezmoi.NullSystem{},
					outputFunc: onepasswordV2OutputFunc,
				},
			}

			require.NotPanics(t, func() {
				actual := c.onepasswordTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["id"], "wxcplh5udshnonkzg2n4qx262y")
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordDetailsFieldsTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["password"].(map[string]interface{})["value"], "L8rm1JXJIE1b8YUDWq7h")
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordDocumentTemplateFunc("ExampleDocument")
				assert.Equal(t, "ExampleDocumentContents", actual)
			})

			require.NotPanics(t, func() {
				actual := c.onepasswordItemFieldsTemplateFunc("ExampleLogin")
				assert.Equal(t, actual["exampleLabel"].(map[string]interface{})["value"], "exampleValue")
			})
		})
	}
}

func onepasswordTestCases(t *testing.T) []onepasswordTestCase {
	t.Helper()
	var testCases []onepasswordTestCase
	require.NoError(t, combinator.Generate(&testCases, struct {
		Prompt      []bool
		EnvironFunc []func() []string
	}{
		Prompt: []bool{false, true},
		EnvironFunc: []func() []string{
			func() []string {
				return nil
			},
			func() []string {
				return []string{
					"OP_SESSION_account=session-token",
				}
			},
			func() []string {
				return []string{
					"OP_SESSION_account1=session-token-1",
					"OP_SESSION_account2=session-token-2",
				}
			},
		},
	}))
	return testCases
}

func onepasswordV1OutputFunc(cmd *exec.Cmd) ([]byte, error) {
	switch {
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--version"}):
		return []byte("1.3.0\n"), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "signin", "--raw"}):
		return []byte("session-token\n"), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "get", "item", "ExampleLogin"}):
		fallthrough
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--session", "session-token", "get", "item", "ExampleLogin"}):
		return []byte(`` +
			`{"uuid":"wxcplh5udshnonkzg2n4qx262y","templateUuid":"001","trashed":"N",` +
			`"createdAt":"2020-07-28T13:44:57Z","updatedAt":"2020-07-28T14:27:46Z",` +
			`"changerUuid":"VBDXOA4MPVHONK5IIJVKUQGLXM","itemVersion":2,"vaultUuid":"tscpxgi6s7c662jtqn3vmw4n5a",` +
			`"details":{"fields":[{"designation":"username","name":"username","type":"T","value":"exampleuser"},` +
			`{"designation":"password","name":"password","type":"P","value":"L8rm1JXJIE1b8YUDWq7h"}],` +
			`"notesPlain":"","passwordHistory":[],"sections":[{"name":"linked items","title":"Related Items"},` +
			`{"fields":[{"k":"string","n":"D4328E0846D2461E8E455D7A07B93397","t":"exampleLabel",` +
			`"v":"exampleValue"}],"name":"Section_20E0BD380789477D8904F830BFE8A121","title":""}]},` +
			`"overview":{"URLs":[{"l":"website","u":"https://www.example.com/"}],"ainfo":"exampleuser",` +
			`"pbe":119.083926,"pgrng":true,"ps":100,"tags":[],"title":"ExampleLogin",` +
			`"url":"https://www.example.com/"}}`,
		), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "get", "document", "ExampleDocument"}):
		fallthrough
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--session", "session-token", "get", "document", "ExampleDocument"}):
		return []byte(`ExampleDocumentContents`), nil
	default:
		timeStr := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(cmd.Stderr, "[ERROR] %s unknown command \"%s\" for \"op\"\n", timeStr, strings.Join(cmd.Args[1:], " "))
		return nil, &exec.ExitError{}
	}
}

func onepasswordV2OutputFunc(cmd *exec.Cmd) ([]byte, error) {
	switch {
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--version"}):
		return []byte("2.0.0-beta8\n"), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "signin", "--raw"}):
		return []byte("session-token\n"), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "item", "get", "--format", "json", "ExampleLogin"}):
		fallthrough
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--session", "session-token", "item", "get", "--format", "json", "ExampleLogin"}):
		return []byte(`` +
			`{"id":"wxcplh5udshnonkzg2n4qx262y","title":"ExampleLogin","version":1,"vault":` +
			`{"id":"tscpxgi6s7c662jtqn3vmw4n5a"},"category":"LOGIN","last_edited_by":"YO4UTYPAD3ZFBNZG5DVAZFBNZM",` +
			`"created_at":"2022-01-17T01:53:50Z","updated_at":"2022-01-17T01:55:35Z",` +
			`"sections":[{"id":"Section_cdzjhg2jo7jylpyin2f5mbfnhm","label":"Related Items"}],` +
			`"fields":[{"id":"username","type":"STRING","purpose":"USERNAME","label":"username",` +
			`"value":"exampleuser"},{"id":"password","type":"CONCEALED","purpose":"PASSWORD","label":"password",` +
			`"value":"L8rm1JXJIE1b8YUDWq7h","password_details":{"strength":"EXCELLENT"}},{"id":"notesPlain",` +
			`"type":"STRING","purpose":"NOTES","label":"notesPlain"},{"id":"cqn7oda7wkcsar7rzcr52i2m3u",` +
			`"section":{"id":"Section_cdzjhg2jo7jylpyin2f5mbfnhm","label":"Related Items"},"type":"STRING",` +
			`"label":"exampleLabel","value":"exampleValue"}],"urls":[{"primary":true,` +
			`"href":"https://www.example.com/"}]}`,
		), nil
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "document", "get", "ExampleDocument"}):
		fallthrough
	case assert.ObjectsAreEqual(cmd.Args, []string{"op", "--session", "session-token", "document", "get", "ExampleDocument"}):
		return []byte(`ExampleDocumentContents`), nil
	default:
		timeStr := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(cmd.Stderr, "[ERROR] %s unknown command \"%s\" for \"op\"\n", timeStr, strings.Join(cmd.Args[1:], " "))
		return nil, &exec.ExitError{}
	}
}
