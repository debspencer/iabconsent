package iabconsent_test

import (
	"github.com/go-check/check"

	"github.com/debspencer/iabconsent"
)

type ParsedConsentSuite struct{}

func (p *ParsedConsentSuite) TestParseConsentStrings(c *check.C) {
	var cases = []struct {
		Type          consentType
		EncodedString string
	}{
		{
			Type:          BitField,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAplY",
		},
		{
			Type:          SingleRangeWithSingleID,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAqABAD2AAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Type:          SingleRangeWithRange,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAqABgD2AdQAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Type:          MultipleRangesWithSingleID,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAqACAD2AOoAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Type:          MultipleRangesWithRange,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAqACgD2AdUBWQHIAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Type:          MultipleRangesMixed,
			EncodedString: "BONMj34ONMj34ABACDENALqAAAAAqACAD3AVkByAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Type:          V2,
			EncodedString: "COvVNSUOvVNSUKyACDENAPCEANAAABwAAAIgBAwAgAVQCAAIEAgYAQAQoBAAECAA.IFoEUQQgAIQwgIwQABAEAAAAOIAACAIAAAAQAIAgEAACEAAAAAgAQBAAAAAAAGBAAgAAAAAAAFAAECAAAgAAQARAEQAAAAAJAAIAAgAAAYQEAAAQmAgBC3ZAYzUw.QD5QAoBAAECAfIA",
		},
	}

	for _, tc := range cases {
		c.Log(tc)
		pc, err := iabconsent.Parse(tc.EncodedString)
		c.Check(err, check.IsNil)

		gotAllowedVendors := pc.AllowedVendors
		gotDisclosedVendors := pc.DisclosedVendors
		gotPublisherTC := pc.PublisherTC

		pc.AllowedVendors = nil
		pc.DisclosedVendors = nil
		pc.PublisherTC = nil

		expect := consentFixtures[tc.Type]

		expectAllowedVendors := expect.AllowedVendors
		expectDisclosedVendors := expect.DisclosedVendors
		expectPublisherTC := expect.PublisherTC

		expect.AllowedVendors = nil
		expect.DisclosedVendors = nil
		expect.PublisherTC = nil

		c.Assert(pc, check.DeepEquals, expect)

		c.Assert(gotAllowedVendors, check.DeepEquals, expectAllowedVendors)
		c.Assert(gotDisclosedVendors, check.DeepEquals, expectDisclosedVendors)
		c.Assert(gotPublisherTC, check.DeepEquals, expectPublisherTC)
	}
}

var _ = check.Suite(&ParsedConsentSuite{})
