package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripCDATA(t *testing.T) {
	tests := []struct {
		str string
		res string
	}{
		{"<![CDATA[ test ]]>test", " test test"},
		{"<![CDATA[test &]]> &lt;", "test & <"},
		{"", ""},
		{"test", "test"},
		{"]]>", "]]>"},
		{"<![CDATA[", "<![CDATA["},
		{"<![CDATA[testtest", "<![CDATA[testtest"},
		{`<![CDATA[
    Since this is a CDATA section
    I can use all sorts of reserved characters
    like > < " and &
    or write things like
    <foo></bar>
    but my document is still well formed!
]]>`, `
    Since this is a CDATA section
    I can use all sorts of reserved characters
    like > < " and &
    or write things like
    <foo></bar>
    but my document is still well formed!
`},
		{`<![CDATA[
Within this Character Data block I can
use double dashes as much as I want (along with <, &, ', and ")
*and* %MyParamEntity; will be expanded to the text
"Has been expanded" ... however, I can't use
the CEND sequence. If I need to use CEND I must escape one of the
brackets or the greater-than sign using concatenated CDATA sections.
]]>`, `
Within this Character Data block I can
use double dashes as much as I want (along with <, &, ', and ")
*and* %MyParamEntity; will be expanded to the text
"Has been expanded" ... however, I can't use
the CEND sequence. If I need to use CEND I must escape one of the
brackets or the greater-than sign using concatenated CDATA sections.
`},
		// 		{`<![CDATA[ test ]]><!--
		// Within this comment I can use ]]>
		// and other reserved characters like <
		// &, ', and ", but %MyParamEntity; will not be expanded
		// (if I retrieve the text of this node it will contain
		// %MyParamEntity; and not "Has been expanded")
		// and I can't place two dashes next to each other.
		// -->`, ` test <!--
		// Within this comment I can use ]]>
		// and other reserved characters like <
		// &, ', and ", but %MyParamEntity; will not be expanded
		// (if I retrieve the text of this node it will contain
		// %MyParamEntity; and not "Has been expanded")
		// and I can't place two dashes next to each other.
		// -->`,
		// 		},
		{`<![CDATA[ test ]]><!-- test -->`, ` test <!-- test -->`}, // TODO: probably wrong
		{`An example of escaped CENDs`, `An example of escaped CENDs`},
		{`<![CDATA[This text contains a CEND ]]]]><![CDATA[>]]>`, `This text contains a CEND ]]>`},
		{`<![CDATA[This text contains a CEND ]]]><![CDATA[]>]]>`, `This text contains a CEND ]]>`},
	}

	for _, test := range tests {
		res := StripCDATA(test.str)
		assert.Equal(t, test.res, res)
	}
}
