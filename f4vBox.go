package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

/*
 * F4V formate, see "ADOBE FLASH VIDEO FILE FORMAT SPECIFICATION VERSION 10.1"
 */
type BoxHeader struct {
	TotalSize    uint32
	BoxType      [4]byte
	ExtendedSize uint64
}

type FragmentRunEntry struct {
	FirstFragment          uint32
	FirstFragmentTimestamp uint64
	FragmentDuration       uint32
	DiscontinuityIndicator uint8 //IF FragmentDuration == 0 UI8
}

// Fragment Run Table
type AfrtBox struct {
	Header BoxHeader // BoxType ='afrt' (0x61667274)

	Version uint8 // UI8, Either 0 or 1

	Flags [3]byte // UI24, The following values are defined: 0 = A full table.
	// 1 = The records in this table are updates (or new entries to be appended) to
	// the previously defined Fragment Run Table. The Update flag in
	// the containing Bootstrap Info box shall be 1 when this flag is 1.

	TimeScale uint32 // UI32, The number of time units per second, used in the FirstFragmentTimestamp and FragmentDuration fields. Typically, the value is 1.

	QualityEntryCount uint8 // UI8, The number of QualitySegmentUrlModifiers (quality level references) that follow.
	// If 0, this Fragment Run Table applies to all quality levels, and there shall be
	// only one Fragment Run Table box in the Bootstrap Info box.

	QualitySegmentUrlModifiers []string

	FragmentRunEntryCount uint32 // UI32, The number of items in this FragmentRunEntryTable. The minimum value is 1.

	FragmentRunEntryTable []FragmentRunEntry
}

func ParseAfrt(data []byte) (*AfrtBox, error) {
	afrtBox := new(AfrtBox)
	parsed := 0
	if len(data) < 8 {
		return nil, errors.New("Box size is smaller than 8 byte")
	}
	// parse header
	afrtBox.Header.TotalSize = binary.BigEndian.Uint32(data[:4])
	parsed += 4
	copy(afrtBox.Header.BoxType[:], data[parsed:parsed+4])
	parsed += 4
	if afrtBox.Header.TotalSize == 1 {
		afrtBox.Header.ExtendedSize = binary.BigEndian.Uint64(data[parsed : parsed+8])
		parsed += 8
	}
	//fmt.Printf("Box size = %d, Box type = %s\n", afrtBox.Header.TotalSize, string(afrtBox.Header.BoxType[:]))

	// parse paylaod
	afrtBox.Version = uint8(data[parsed])
	parsed++
	copy(afrtBox.Flags[:], data[parsed:parsed+3])
	parsed += 3
	afrtBox.TimeScale = binary.BigEndian.Uint32(data[parsed : parsed+4])
	parsed += 4
	afrtBox.QualityEntryCount = uint8(data[parsed])
	parsed++
	afrtBox.QualitySegmentUrlModifiers = make([]string, afrtBox.QualityEntryCount)
	for i := 0; i < int(afrtBox.QualityEntryCount); i++ {
		for {
			afrtBox.QualitySegmentUrlModifiers[i] += string(data[parsed])
			if int(data[parsed]) == 0 { // null-terminate
				parsed++
				break
			}
			parsed++
		}
	}
	afrtBox.FragmentRunEntryCount = binary.BigEndian.Uint32(data[parsed : parsed+4])
	parsed += 4
	afrtBox.FragmentRunEntryTable = make([]FragmentRunEntry, afrtBox.FragmentRunEntryCount)
	for i := 0; i < int(afrtBox.FragmentRunEntryCount); i++ {
		afrtBox.FragmentRunEntryTable[i].FirstFragment = binary.BigEndian.Uint32(data[parsed : parsed+4])
		parsed += 4
		afrtBox.FragmentRunEntryTable[i].FirstFragmentTimestamp = binary.BigEndian.Uint64(data[parsed : parsed+8])
		parsed += 8
		afrtBox.FragmentRunEntryTable[i].FragmentDuration = binary.BigEndian.Uint32(data[parsed : parsed+4])
		parsed += 4
		if int(afrtBox.FragmentRunEntryTable[i].FragmentDuration) == 0 {
			afrtBox.FragmentRunEntryTable[i].DiscontinuityIndicator = uint8(data[parsed])
			parsed++
		}
	}
	return afrtBox, nil
}

type SegmentRunEntry struct {
	FirstSegment uint32 // UI32,The identifying number of the first segment in the run of
	// segments containing the same number of fragments.
	// The segment corresponding to the FirstSegment in
	// the next SEGMENTRUNENTRY will terminate this run.

	FragmentsPerSegment uint32 // UI32,The number of fragments in each segment in this run.
}

// Segment Run Table
type AsrtBox struct {
	Header BoxHeader // BoxType = 'asrt' (0x61737274)

	Version uint8 // UI8,Either 0 or 1

	Flags [3]byte // UI24,The following values are defined: 0 = A full table.
	// 1 = The records in this table are updates (or new entries to be appended) to
	// the previously defined Segment Run Table.
	// The Update flag in the containing Bootstrap Info box shall be 1 when this flag is 1

	QualityEntryCount uint8 // UI8,The number of QualitySegmentUrlModifiers (quality level references) that follow.
	// If 0, this Segment Run Table applies to all quality levels,
	//and there shall be only one Segment Run Table box in the Bootstrap Info box.

	QualitySegmentUrlModifiers []string

	SegmentRunEntryCount uint32 // The number of items in this SegmentRunEntryTable. The minimum value is 1.

	SegmentRunEntryTable []SegmentRunEntry // Array of segment run entries
}

func ParseAsrt(data []byte) (*AsrtBox, error) {
	asrtBox := new(AsrtBox)
	parsed := 0
	if len(data) < 8 {
		return nil, errors.New("Box size is smaller than 8 byte")
	}
	// parse header
	asrtBox.Header.TotalSize = binary.BigEndian.Uint32(data[:4])
	parsed += 4
	copy(asrtBox.Header.BoxType[:], data[parsed:parsed+4])
	parsed += 4
	if asrtBox.Header.TotalSize == 1 {
		asrtBox.Header.ExtendedSize = binary.BigEndian.Uint64(data[parsed : parsed+8])
		parsed += 8
	}
	// fmt.Printf("Box size = %d, Box type = %s\n", asrtBox.Header.TotalSize, string(asrtBox.Header.BoxType[:]))

	// parse payload
	asrtBox.Version = uint8(data[parsed])
	parsed++
	copy(asrtBox.Flags[:], data[parsed:parsed+3])
	parsed += 3
	asrtBox.QualityEntryCount = uint8(data[parsed])
	parsed++
	asrtBox.QualitySegmentUrlModifiers = make([]string, asrtBox.QualityEntryCount)
	for i := 0; i < int(asrtBox.QualityEntryCount); i++ {
		for {
			asrtBox.QualitySegmentUrlModifiers[i] += string(data[parsed])
			if int(data[parsed]) == 0 { // nil-terminated
				parsed++
				break
			}
			parsed++
		}
	}
	asrtBox.SegmentRunEntryCount = binary.BigEndian.Uint32(data[parsed : parsed+4])
	parsed += 4
	asrtBox.SegmentRunEntryTable = make([]SegmentRunEntry, asrtBox.SegmentRunEntryCount)
	for i := 0; i < int(asrtBox.SegmentRunEntryCount); i++ {
		asrtBox.SegmentRunEntryTable[i].FirstSegment = binary.BigEndian.Uint32(data[parsed : parsed+4])
		parsed += 4
		asrtBox.SegmentRunEntryTable[i].FragmentsPerSegment = binary.BigEndian.Uint32(data[parsed : parsed+4])
		parsed += 4
	}
	return asrtBox, nil
}

type QualityEntry struct {
	QualitySegmentUrlModifier string
}

type ServerEntry struct {
	ServerBaseURL string
}

// Bootstrap Info (abst) box
type AbstBox struct {
	Header BoxHeader // BOXHEADER, BoxType = 'abst' (0x61627374)

	Version uint8 //UI8, Either 0 or 1

	Flags [3]byte // UI24, Reserved. Set to 0

	BootstrapinfoVersion uint32 // UI32, The version number of the bootstrap information.
	// When the Update field is set, BootstrapinfoVersion indicates the version number that is being updated.

	Profile int // UI2, Indicates if it is the Named Access (0) or the Range Access (1) Profile.
	// One bit is reserved for future profiles.

	Live int // UI1, Indicates if the media presentation is live (1) or not.

	Update int // UI1, Indicates if this table is a full version (0) or an update (1) to a previously defined (sent) full version of the bootstrap box or file.
	// Updates are not complete replacements. They may contain only the changed elements. The server sends the updates only when the bootstrap information changes.
	// The updates apply to the full version with the same BootstrapinfoVersion number. There may be more than one update for the same BootstrapinfoVersion number.
	// If the server sends multiple updates, the updates apply to the full version with the same BootstrapinfoVersion number.
	// Each update includes all previous updates to the same BootstrapinfoVersion.
	// For multiple updates to a single full version, the latest update is determined based on the CurrentMediaTime.

	Reserved int // UI4, Reserved, set to 0

	TimeScale uint32 // UI32, The number of time units per second.
	// The field CurrentMediaTime uses this value to represent accurate time.
	// Typically, the value is 1000, for a unit of milliseconds.

	CurrentMediaTime uint64 // UI64,The timestamp in TimeScale units of the latest available Fragment in the media presentation.
	// This timestamp is used to request the right fragment number.
	// The CurrentMediaTime can be the total duration.
	// For media presentations that are not live, CurrentMediaTime can be 0.

	SmpteTimeCodeOffset uint64 // UI64, The offset of the CurrentMediaTime from the SMPTE time code,
	//  converted to milliseconds. This offset is not in TimeScale units.
	// This field is zero when not used.
	//The server uses the SMPTE time code modulo 24 hours to make the offset positive.

	MovieIdentifier string // string, The identifier of this presentation.
	// The identifier is a null-terminated UTF-8 string.
	// For example, it can be a filename or pathname in a URL.

	ServerEntryCount uint8 // UI8, The number of ServerEntryTable entries. The minimum value is 0.

	ServerEntryTable []ServerEntry // SERVERENTRY[ServerEntryCount],Server URLs in descending order of preference
	// ServerBaseURL STRING The server base url for this presentation on that server.
	// The value is a null-terminated UTF-8 string, without a trailing "/".

	QualityEntryCount uint8 // UI8, The number of QualityEntryTable entries, which is also the number of available quality levels.
	// The minimum value is 0. Available quality levels are for, for example, multi bit rate files or trick files.

	QualityEntryTable []QualityEntry // Quality file references in order from high to low quality

	DrmData string // Null or null-terminated UTF-8 string.
	// This string holds Digital Rights Management metadata.
	// Encrypted files use this metadata to get the necessary keys and licenses for decryption and playback.
	MetaData string // Null or null-terminated UTF-8 string that holds metadata

	SegmentRunTableCount uint8 // UI8 The number of entries in SegmentRunTableEntries.(asrt)
	// The minimum value is 1. Typically, one table contains all segment runs.
	// However, this count provides the flexibility to define the segment runs individually for each quality level (or trick file).

	SegmentRunTableEntries []*AsrtBox // Array of SegmentRunTable elements

	FragmentRunTableCount uint8 // UI8, The number of entries in FragmentRunTable-Entries. The minimum value is 1.

	FragmentRunTableEntries []*AfrtBox // FragmentRunTable[FragmentRunTableCount], Array of FragmentRunTable elements
}

func ParseAbst(data []byte) (*AbstBox, error) {
	abstBox := new(AbstBox)
	parsed := 0
	if len(data) < 8 {
		return nil, errors.New("Box size is smaller than 8 byte")
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[ERROR] Error happened while parse abst\n")
		}
	}()

	// parse header
	abstBox.Header.TotalSize = binary.BigEndian.Uint32(data[:4])
	parsed += 4
	copy(abstBox.Header.BoxType[:], data[parsed:parsed+4])
	parsed += 4
	if abstBox.Header.TotalSize == 1 {
		abstBox.Header.ExtendedSize = binary.BigEndian.Uint64(data[parsed : parsed+8])
		parsed += 8
	}
	// fmt.Printf("Box size = %d, Box type = %s\n", abstBox.Header.TotalSize, string(abstBox.Header.BoxType[:]))

	// parse payload
	payload := data[:]
	abstBox.Version = uint8(payload[parsed])
	parsed++
	copy(abstBox.Flags[:], payload[parsed:parsed+3])
	parsed += 3
	abstBox.BootstrapinfoVersion = binary.BigEndian.Uint32(payload[parsed : parsed+4])
	parsed += 4
	abstBox.Profile = int((payload[parsed] >> 6) & 0x03)
	abstBox.Live = int((payload[parsed] >> 5) & 0x01)
	abstBox.Update = int((payload[parsed] >> 4) & 0x01)
	abstBox.Reserved = int((payload[parsed]) & 0x0f)
	parsed += 1
	abstBox.TimeScale = binary.BigEndian.Uint32(payload[parsed : parsed+4])
	parsed += 4
	abstBox.CurrentMediaTime = binary.BigEndian.Uint64(payload[parsed : parsed+8])
	parsed += 8
	abstBox.SmpteTimeCodeOffset = binary.BigEndian.Uint64(payload[parsed : parsed+8])
	parsed += 8

	for {
		abstBox.MovieIdentifier += string(payload[parsed])
		if int(payload[parsed]) == 0 { // null-terminated
			parsed++
			break
		}
		parsed++
	}

	abstBox.ServerEntryCount = uint8(payload[parsed])
	parsed++
	abstBox.ServerEntryTable = make([]ServerEntry, abstBox.ServerEntryCount)
	// fmt.Printf("EndtryCount = %d, parsed = %d, Identifier = %s\n", abstBox.ServerEntryCount, parsed, abstBox.MovieIdentifier)
	for i := 0; i < int(abstBox.ServerEntryCount); i++ {
		for {
			abstBox.ServerEntryTable[i].ServerBaseURL += string(payload[parsed])
			if int(payload[parsed]) == 0 { // null-terminated
				parsed++
				break
			}
			parsed++
		}
	}

	abstBox.QualityEntryCount = uint8(payload[parsed])
	parsed++
	abstBox.QualityEntryTable = make([]QualityEntry, abstBox.QualityEntryCount)
	for i := 0; i < int(abstBox.QualityEntryCount); i++ {
		for {
			abstBox.QualityEntryTable[i].QualitySegmentUrlModifier += string(payload[parsed])
			if int(payload[parsed]) == 0 { // null-terminated
				parsed++
				break
			}
			parsed++
		}
	}

	for {
		abstBox.DrmData += string(payload[parsed])
		if int(payload[parsed]) == 0 { //null-terminate
			parsed++
			break
		}
		parsed++
	}

	for {
		abstBox.MetaData += string(payload[parsed])
		if int(payload[parsed]) == 0 {
			parsed++
			break
		}
		parsed++
	}

	abstBox.SegmentRunTableCount = uint8(payload[parsed])
	parsed++
	abstBox.SegmentRunTableEntries = make([]*AsrtBox, abstBox.SegmentRunTableCount)
	for i := 0; i < int(abstBox.SegmentRunTableCount); i++ {
		segmentRunTableEntrie, err := ParseAsrt(payload[parsed:])
		if err != nil {
			log.Println("Parse asrt failed:", err.Error())
			return nil, errors.New("Parse asrt failed")
		}
		abstBox.SegmentRunTableEntries[i] = segmentRunTableEntrie
		parsed += int(abstBox.SegmentRunTableEntries[i].Header.TotalSize)
	}

	abstBox.FragmentRunTableCount = uint8(payload[parsed])
	parsed++
	abstBox.FragmentRunTableEntries = make([]*AfrtBox, abstBox.FragmentRunTableCount)
	for i := 0; i < int(abstBox.FragmentRunTableCount); i++ {
		fragmentRunTableEntrie, err := ParseAfrt(payload[parsed:])
		if err != nil {
			log.Println("Parse afrt failed", err.Error())
			return nil, errors.New("Parse afrt failed")
		}
		abstBox.FragmentRunTableEntries[i] = fragmentRunTableEntrie
		parsed += int(abstBox.FragmentRunTableEntries[i].Header.TotalSize)
	}

	return abstBox, nil
}

func printHeader(header BoxHeader) {
	fmt.Printf("TotalSize: %d\n", header.TotalSize)
	fmt.Printf("BoxType: %s\n", string(header.BoxType[:]))
	if header.TotalSize == 1 {
		fmt.Printf("ExtendedSize: %d\n", header.ExtendedSize)
	}
}

func (abst *AbstBox) Print() {
	printHeader(abst.Header)
	fmt.Printf("Version: %d\n", abst.Version)
	fmt.Printf("Flags: %v\n", abst.Flags[:])
	fmt.Printf("BootstrapinfoVersion: %d\n", abst.BootstrapinfoVersion)
	fmt.Printf("Profile: %d\n", abst.Profile)
	fmt.Printf("Live: %d\n", abst.Live)
	fmt.Printf("Update: %d\n", abst.Update)
	fmt.Printf("Reserved: %d\n", abst.Reserved)
	fmt.Printf("TimeScale: %d\n", abst.TimeScale)
	fmt.Printf("CurrentMediaTime: %d\n", abst.CurrentMediaTime)
	fmt.Printf("SmpteTimeCodeOffset: %d\n", abst.SmpteTimeCodeOffset)
	fmt.Printf("MovieIdentifier: %s\n", abst.MovieIdentifier)
	fmt.Printf("ServerEntryCount: %d\n", abst.ServerEntryCount)
	for i := 0; i < int(abst.ServerEntryCount); i++ {
		fmt.Printf("ServerEntryTable[%d].ServerBaseUrl: %s\n", i, abst.ServerEntryTable[i].ServerBaseURL)
	}

	fmt.Printf("QualityEntryCount: %d\n", abst.QualityEntryCount)
	for i := 0; i < int(abst.QualityEntryCount); i++ {
		fmt.Printf("QualityEntryTable[%d].QualitySegmentUrlModifier: %s\n", i, abst.QualityEntryTable[i].QualitySegmentUrlModifier)
	}

	fmt.Printf("DrmData: %s\n", abst.DrmData)
	fmt.Printf("MetaData: %s\n", abst.MetaData)
	fmt.Printf("SegmentRunTableCount: %d\n", abst.SegmentRunTableCount)
	for i := 0; i < int(abst.SegmentRunTableCount); i++ {
		fmt.Printf("SegmentRunTableEntries[%d]: \n", i)
		fmt.Println("---ASRT---")
		abst.SegmentRunTableEntries[i].Print()
	}
	fmt.Printf("FragmentRunTableCount: %d\n", abst.FragmentRunTableCount)
	for i := 0; i < int(abst.FragmentRunTableCount); i++ {
		fmt.Printf("FragmentRunTableEntries[%d]: \n", i)
		fmt.Println("---AFRT---")
		abst.FragmentRunTableEntries[i].Print()
	}
}

func (asrt *AsrtBox) Print() {
	printHeader(asrt.Header)
	fmt.Printf("Version: %d\n", asrt.Version)
	fmt.Printf("Flags: %v\n", asrt.Flags[:])
	fmt.Printf("QualityEntryCount: %d\n", asrt.QualityEntryCount)
	for i := 0; i < int(asrt.QualityEntryCount); i++ {
		fmt.Printf("QualitySegmentUrlModifiers: %s\n", asrt.QualitySegmentUrlModifiers)
	}
	fmt.Printf("SegmentRunEntryCount: %d\n", asrt.SegmentRunEntryCount)
	for i := 0; i < int(asrt.SegmentRunEntryCount); i++ {
		fmt.Printf("SegmentRunEntryTable[%d].FirstSegment: %d\n", i, asrt.SegmentRunEntryTable[i].FirstSegment)
		fmt.Printf("SegmentRunEntryTable[%d].FragmentsPerSegment: %d\n", i, asrt.SegmentRunEntryTable[i].FragmentsPerSegment)
	}
}

func (afrt *AfrtBox) Print() {
	printHeader(afrt.Header)
	fmt.Printf("Version: %d\n", afrt.Version)
	fmt.Printf("Flags: %v\n", afrt.Flags[:])
	fmt.Printf("TimeScale: %d\n", afrt.TimeScale)
	fmt.Printf("QualityEntryCount: %d\n", afrt.QualityEntryCount)
	fmt.Printf("QualitySegmentUrlModifiers: %s\n", afrt.QualitySegmentUrlModifiers)
	fmt.Printf("FragmentRunEntryCount: %d\n", afrt.FragmentRunEntryCount)
	for i := 0; i < int(afrt.FragmentRunEntryCount); i++ {
		fmt.Printf("FragmentRunEntryTable[%d].FirstFragment: %d\n", i, afrt.FragmentRunEntryTable[i].FirstFragment)
		fmt.Printf("FragmentRunEntryTable[%d].FirstFragmentTimestamp: %d\n", i, afrt.FragmentRunEntryTable[i].FirstFragmentTimestamp)
		fmt.Printf("FragmentRunEntryTable[%d].FragmentDuration: %d\n", i, afrt.FragmentRunEntryTable[i].FragmentDuration)
		if afrt.FragmentRunEntryTable[i].FragmentDuration == 0 {
			fmt.Printf("FragmentRunEntryTable[%d].DiscontinuityIndicator: %d\n", i, afrt.FragmentRunEntryTable[i].DiscontinuityIndicator)
		}
	}
}
