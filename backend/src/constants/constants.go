package constants

const MojangVersionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
const BedrockVersionAPIURL = "https://raw.githubusercontent.com/Bedrock-OSS/BDS-Versions/master/versions.json"

var BedrockVersions = []string{} // Deprecated, will be fetched from API


const DefaultJavaPort uint16 = 25565
const DefaultBedrockPort uint16 = 19132
const DefaultMaxPlayers uint32 = 20
const DefaultMemoryMB uint32 = 1024
