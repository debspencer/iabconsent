/*

Package iabconsent provides structs and methods for parsing
Vendor Consent Strings as defined by the IAB Consent String 1.1 Spec.
More info on the spec here:
https://github.com/InteractiveAdvertisingBureau/GDPR-Transparency-and-Consent-Framework/blob/master/Consent%20string%20and%20vendor%20list%20formats%20v1.1%20Final.md#vendor-consent-string-format-.

Copyright (c) 2018 LiveRamp. All rights reserved.

Written by Andy Day, Software Engineer @ LiveRamp
for use in the LiveRamp Pixel Server.

*/
package iabconsent

import (
	"time"
)

// ParsedConsent represents data extracted from an IAB Consent String, v1.1.
type ParsedConsent struct {
	Version               int               // 1.1 2.0
	Created               time.Time         // 1.1 2.0
	LastUpdated           time.Time         // 1.1 2.0
	CMPID                 int               // 1.1 2.0
	CMPVersion            int               // 1.1 2.0
	ConsentScreen         int               // 1.1 2.0
	ConsentLanguage       string            // 1.1 2.0
	VendorListVersion     int               // 1.1 2.0
	PolicyVersion         int               //     2.0
	IsSpecificService     bool              //     2.0
	UseNonStandardStacks  bool              //     2.0
	SpecialFeatureOptins  map[int]bool      //     2.0
	PurposesAllowed       map[int]bool      // 1.1 2.0
	PurposesTransparancy  map[int]bool      //     2.0
	PurposeOneTreatment   bool              //     2.0
	PublisherCC           string            //     2.0
	MaxVendorID           int               // 1.1 2.0
	ConsentedVendors      map[int]bool      // 1.1 2.0
	LegitMaxVendorID      int               //     2.0
	LegitConsentedVendors map[int]bool      //     2.0
	NumPubRestrictions    int               //     2.0
	PubRestrictions       []*PubRestriction //     2.0
	DisclosedVendors      *Vendors          //     2.0
	AllowedVendors        *Vendors          //     2.0
	PublisherTC           *PublisherTC      //     2.0
}

type PubRestriction struct {
	PurposeId         int
	RestrictionType   int
	NumEntries        int
	RestrictedVendors map[int]bool
}

type Vendors struct {
	SegmentType      int
	MaxVendorID      int
	ConsentedVendors map[int]bool
}

type PublisherTC struct {
	SegmentType                  int
	PubPurposesConsent           map[int]bool
	PubPurposesLITransparency    map[int]bool
	NumCustomPurposes            int
	CustomPurposesConsent        map[int]bool
	CustomPurposesLITransparency map[int]bool
}

// EveryPurposeAllowed returns true iff every purpose number in ps exists in
// the ParsedConsent, otherwise false.
func (p *ParsedConsent) EveryPurposeAllowed(ps []int) bool {
	for _, rp := range ps {
		if !p.PurposesAllowed[rp] {
			return false
		}
	}
	return true
}

// VendorAllowed returns true if the ParsedConsent contains affirmative consent
// for VendorID v.
func (p *ParsedConsent) VendorAllowed(v int) bool {
	return p.ConsentedVendors[v]
}
