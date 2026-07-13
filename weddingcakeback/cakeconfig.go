package weddingcakeback

// TierTop holds no DonutForests of SingleTree's (it works differently)
/* THIS TABLE WILL LIKELY CHANGE SOON (there will be more smaller tiers)
Tier			Max DonutForests	SingleTree's per DF		Hashes per SingleTree	Total hashes
----			----------------	-------------------    	---------------------	------------
TierTop			N/A					N/A						N/A						65535
TierBelow[0]	255					1 (=256^0)				65,535					~16 million (255 * 65535)
TierBelow[1]	255					256	(=256^1)			~65,535					~4 billion (255 * 256 * 65535)
TierBelow[2]	255					65,536 (=256^2) 		~65,535					~1 trillion (255 * 65536 * 65535)
TierBelow[3]	255					16,777,216 (=256^3)		~65,535					~280 trillion (255 * 16M * 65535)
*/

const MaxTiersBelowCount = 4 // Far more than most PC's can handle!

// CakeConfig holds some parameters of the cake being built
type CakeConfig struct {
	// Global config
	HashLength byte

	// TierTop has no config (it works differently)

	// TierBelow Configs
	TierBelowConfigs [MaxTiersBelowCount]TierBelowConfig
}

// Mostly hard coded for now
func NewCakeConfig(hashLength byte, reassuranceBytes byte) *CakeConfig {
	result := CakeConfig{}
	result.HashLength = hashLength
	for tierBelow := 0; tierBelow < MaxTiersBelowCount; tierBelow++ {
		result.TierBelowConfigs[tierBelow].NodeFormatSpecsPerLevel = 10
		result.TierBelowConfigs[tierBelow].ReassuranceBytesCount = reassuranceBytes
	}
	result.TierBelowConfigs[0].MaxDonutForests = 255
	result.TierBelowConfigs[0].NodeIdConfig = ID16[NodeIdType]{}           // 16 bits per node id
	result.TierBelowConfigs[0].HashIndexIdConfig = ID16[HashIndexIdType]{} // 16 bits per hash index id
	result.TierBelowConfigs[1].MaxDonutForests = 255
	result.TierBelowConfigs[1].NodeIdConfig = ID24[NodeIdType]{}           // 24 bits per node id
	result.TierBelowConfigs[1].HashIndexIdConfig = ID24[HashIndexIdType]{} // 24 bits per hash index id
	result.TierBelowConfigs[2].MaxDonutForests = 255
	result.TierBelowConfigs[2].NodeIdConfig = ID32[NodeIdType]{}           // 32 bits per node id
	result.TierBelowConfigs[2].HashIndexIdConfig = ID32[HashIndexIdType]{} // 32 bits per hash index id
	result.TierBelowConfigs[3].MaxDonutForests = 255
	result.TierBelowConfigs[3].NodeIdConfig = ID40[NodeIdType]{}           // 40 bits per node id
	result.TierBelowConfigs[3].HashIndexIdConfig = ID40[HashIndexIdType]{} // 40 bits per hash index id

	return &result
}

type TierBelowConfig struct {
	NodeFormatSpecsPerLevel byte
	ReassuranceBytesCount   byte
	MaxDonutForests         byte
	NodeIdConfig            NByteIdConfig[NodeIdType]
	HashIndexIdConfig       NByteIdConfig[HashIndexIdType]
}
