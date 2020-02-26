package iabconsent

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rupertchen/go-bits"
)

const (
	// dsPerS is deciseconds per second
	dsPerS = 10
	// nsPerDs is nanoseconds per decisecond
	nsPerDs = int64(time.Millisecond * 100)
)

var (
	ErrTooShort           = errors.New("Consent String is too short")
	ErrUnsupportedVersion = errors.New("Unsupport version")
)

// ConsentReader provides additional Consent String-specific bit-reading
// functionality on top of bits.Reader.
type ConsentReader struct {
	*bits.Reader
}

// NewConsentReader returns a new ConsentReader backed by src.
func NewConsentReader(src []byte) *ConsentReader {
	return &ConsentReader{bits.NewReader(bits.NewBitmap(src))}
}

// ReadInt reads the next n bits and converts them to an int.
func (r *ConsentReader) ReadInt(n uint) (int, error) {
	if b, err := r.ReadBits(n); err != nil {
		return 0, errors.WithMessage(err, "read int")
	} else {
		return int(b), nil
	}

}

// ReadTime reads the next 36 bits representing the epoch time in deciseconds
// and converts it to a time.Time.
func (r *ConsentReader) ReadTime() (time.Time, error) {
	if b, err := r.ReadBits(36); err != nil {
		return time.Time{}, errors.WithMessage(err, "read time")
	} else {
		var ds = int64(b)
		return time.Unix(ds/dsPerS, (ds%dsPerS)*nsPerDs).UTC(), nil
	}
}

// ReadString returns a string of length n by reading the next 6 * n bits.
func (r *ConsentReader) ReadString(n uint) (string, error) {
	var buf = make([]byte, 0, n)
	for i := uint(0); i < n; i++ {
		if b, err := r.ReadBits(6); err != nil {
			return "", errors.WithMessage(err, "read string")
		} else {
			buf = append(buf, byte(b)+'A')
		}
	}
	return string(buf), nil
}

// ReadBitField reads the next n bits and converts them to a map[int]bool.
func (r *ConsentReader) ReadBitField(n uint) (map[int]bool, error) {
	var m = make(map[int]bool)
	for i := uint(0); i < n; i++ {
		if b, err := r.ReadBool(); err != nil {
			return nil, errors.WithMessage(err, "read bit field")
		} else {
			if b {
				m[int(i)+1] = true
			}
		}
	}
	return m, nil
}

func (r *ConsentReader) ReadRangeEntries(n uint, max int, def bool) (map[int]bool, error) {
	var ret = make(map[int]bool)

	// if the default is true, the set everything to true so that it can be piecemeal set to false below
	if def == true {
		for i := 0; i <= max; i++ {
			ret[i] = def
		}
	}
	def = !def // invert default for those we do find

	var err error
	for i := uint(0); i < n; i++ {
		var isRange bool
		if isRange, err = r.ReadBool(); err != nil {
			return nil, errors.WithMessage(err, "is-range check")
		}
		var start, end int
		if start, err = r.ReadInt(16); err != nil {
			return nil, errors.WithMessage(err, "range start")
		}
		if isRange {
			if end, err = r.ReadInt(16); err != nil {
				return nil, errors.WithMessage(err, "range end")
			}
		} else {
			end = start
		}
		for n := start; n <= end; n++ {
			ret[n] = def
		}
	}
	return ret, nil
}

func (r *ConsentReader) ReadPubRestrictions(n uint) ([]*PubRestriction, error) {
	var ret = make([]*PubRestriction, 0, n)
	var err error
	for i := uint(0); i < n; i++ {
		p := &PubRestriction{}
		p.PurposeId, err = r.ReadInt(16)
		if err != nil {
			return nil, errors.WithMessage(err, "purpose id")
		}
		p.RestrictionType, err = r.ReadInt(2)
		if err != nil {
			return nil, errors.WithMessage(err, "restriction id")
		}
		p.NumEntries, err = r.ReadInt(12)
		if err != nil {
			return nil, errors.WithMessage(err, "purpose num entries")
		}
		p.RestrictedVendors, err = r.ReadRangeEntries(uint(p.NumEntries), 0, false)

		ret = append(ret, p)
	}
	return ret, nil
}

// Parse takes a base64 Raw URL Encoded string which represents a Vendor
// Consent String and returns a ParsedConsent with its fields populated with
// the values stored in the string.
//
// Example Usage:
//
//   var pc, err = iabconsent.Parse("BONJ5bvONJ5bvAMAPyFRAL7AAAAMhuqKklS-gAAAAAAAAAAAAAAAAAAAAAAAAAA")
func Parse(s string) (*ParsedConsent, error) {
	if len(s) < 1 {
		return nil, ErrTooShort
	}

	version := 0
	var consentString string
	var consentArray []string

	switch s[0] {
	case 'B':
		version = 1
		consentString = s
	case 'C':
		version = 2
		consentArray = strings.Split(s, ".")
		consentString = consentArray[0]
	default:
		return nil, ErrUnsupportedVersion
	}

	var b, err = base64.RawURLEncoding.DecodeString(consentString)
	if err != nil {
		return nil, errors.Wrap(err, "parse consent string")
	}

	var r = NewConsentReader(b)

	// This block of code directly describes the format of the payload.
	var p = &ParsedConsent{}
	p.Version, _ = r.ReadInt(6)
	p.Created, _ = r.ReadTime()
	p.LastUpdated, _ = r.ReadTime()
	p.CMPID, _ = r.ReadInt(12)
	p.CMPVersion, _ = r.ReadInt(12)
	p.ConsentScreen, _ = r.ReadInt(6)
	p.ConsentLanguage, _ = r.ReadString(2)
	p.VendorListVersion, _ = r.ReadInt(12)

	if version == 2 {
		p.PolicyVersion, _ = r.ReadInt(6)
		p.IsSpecificService, _ = r.ReadBool()
		p.UseNonStandardStacks, _ = r.ReadBool()
		p.SpecialFeatureOptins, _ = r.ReadBitField(12)
	}

	p.PurposesAllowed, _ = r.ReadBitField(24)

	if version == 2 {
		p.PurposesTransparancy, _ = r.ReadBitField(24)
		p.PurposeOneTreatment, _ = r.ReadBool()
		p.PublisherCC, _ = r.ReadString(2)
	}

	p.MaxVendorID, _ = r.ReadInt(16)

	isRangeEncoding, _ := r.ReadBool()
	if isRangeEncoding {
		defaultConsent := false
		if version == 1 {
			defaultConsent, _ = r.ReadBool()
		}
		numEntries, _ := r.ReadInt(12)
		p.ConsentedVendors, _ = r.ReadRangeEntries(uint(numEntries), p.MaxVendorID, defaultConsent)
	} else {
		p.ConsentedVendors, _ = r.ReadBitField(uint(p.MaxVendorID))
	}

	if version == 2 {
		p.LegitMaxVendorID, _ = r.ReadInt(16)
		isRangeEncoding, _ := r.ReadBool()
		if isRangeEncoding {
			numEntries, _ := r.ReadInt(12)
			p.LegitConsentedVendors, _ = r.ReadRangeEntries(uint(numEntries), p.LegitMaxVendorID, false)
		} else {
			p.LegitConsentedVendors, _ = r.ReadBitField(uint(p.MaxVendorID))
		}

		p.NumPubRestrictions, _ = r.ReadInt(12)
		p.PubRestrictions, _ = r.ReadPubRestrictions(uint(p.NumPubRestrictions))

		for i := 1; i < len(consentArray); i++ {
			s := consentArray[i]
			if len(s) < 1 {
				continue
			}
			switch s[0] {
			case 'I': // base64 001000
				p.DisclosedVendors, _ = ParseVendors(s)
			case 'Q': // base64 010000
				p.AllowedVendors, _ = ParseVendors(s)
			case 'Y': // base64 011000
				p.PublisherTC, _ = ParsePublisherTC(s)
			default:
				continue
			}
		}
	}

	return p, r.Err
}

func ParseVendors(s string) (*Vendors, error) {
	var b, err = base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "parse disclosed vendors")
	}

	var r = NewConsentReader(b)

	// This block of code directly describes the format of the payload.
	var p = &Vendors{}
	p.SegmentType, _ = r.ReadInt(3)
	p.MaxVendorID, _ = r.ReadInt(16)
	isRangeEncoding, _ := r.ReadBool()
	if isRangeEncoding {
		numEntries, _ := r.ReadInt(12)
		p.ConsentedVendors, _ = r.ReadRangeEntries(uint(numEntries), p.MaxVendorID, false)
	} else {
		p.ConsentedVendors, _ = r.ReadBitField(uint(p.MaxVendorID))
	}
	return p, r.Err
}

func ParsePublisherTC(s string) (*PublisherTC, error) {
	var b, err = base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "parse disclosed vendors")
	}

	var r = NewConsentReader(b)

	// This block of code directly describes the format of the payload.
	var p = &PublisherTC{}
	p.SegmentType, _ = r.ReadInt(3)
	p.PubPurposesConsent, _ = r.ReadBitField(24)
	p.PubPurposesLITransparency, _ = r.ReadBitField(24)
	p.NumCustomPurposes, _ = r.ReadInt(6)
	p.CustomPurposesConsent, _ = r.ReadBitField(uint(p.NumCustomPurposes))
	p.CustomPurposesLITransparency, _ = r.ReadBitField(uint(p.NumCustomPurposes))

	return p, r.Err
}
