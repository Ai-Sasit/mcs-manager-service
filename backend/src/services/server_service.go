package services

import (
	"encoding/json"
	"fmt"
	"io"
	"mc-manage-backend/src/constants"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

type CreateServerParams struct {
	Name       string
	Edition    models.ServerEdition
	ServerType string
	ModLoader  string
	Version    string
	Port       uint16
	MaxPlayers uint32
	MemoryMB   uint32
}

type SetupProgressFunc func(SetupEvent)

func CreateServer(state *AppState, params CreateServerParams) (*models.ServerConfig, error) {
	return createServer(state, params, nil)
}

func CreateServerWithProgress(state *AppState, params CreateServerParams, progress SetupProgressFunc) (*models.ServerConfig, error) {
	return createServer(state, params, progress)
}

func createServer(state *AppState, params CreateServerParams, progress SetupProgressFunc) (*models.ServerConfig, error) {
	emit := func(step, status, message string, percent int, serverID ...string) {
		if progress == nil {
			return
		}
		event := SetupEvent{
			Type:    "setup",
			Step:    step,
			Status:  status,
			Message: message,
			Percent: percent,
		}
		if len(serverID) > 0 {
			event.ServerID = serverID[0]
		}
		progress(event)
	}

	emit("validate", "active", "Validating server settings", 5)
	id := uuid.New().String()
	serverDir := filepath.Join(state.DataDir, "servers", id)
	logger.Info(fmt.Sprintf("[CreateServer] id=%s name=%s edition=%s version=%s", id, params.Name, params.Edition, params.Version), nil)

	emit("directory", "active", "Creating server directory", 12)
	if err := os.MkdirAll(serverDir, os.ModePerm); err != nil {
		logger.Error("[CreateServer] Failed to create dir: "+err.Error(), nil)
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}
	emit("directory", "success", "Server directory ready", 18)

	// Download server files
	var err error
	serverType := params.ServerType
	if serverType == "" {
		serverType = "vanilla"
	}
	logger.Info(fmt.Sprintf("[CreateServer] Downloading %s server %s type=%s", params.Edition, params.Version, serverType), nil)
	emit("metadata", "active", "Resolving version metadata", 28)
	switch params.Edition {
	case models.EditionJava:
		switch serverType {
		case "paper":
			emit("download", "active", "Downloading Paper server jar", 42)
			err = DownloadPaperServer(params.Version, serverDir)
		case "spigot":
			emit("download", "active", "Downloading Spigot-compatible server jar", 42)
			err = DownloadSpigotServer(params.Version, serverDir)
		case "forge":
			emit("download", "active", "Downloading Forge server", 42)
			err = DownloadForgeServer(params.Version, serverDir)
		case "fabric":
			emit("download", "active", "Downloading Fabric server", 42)
			err = DownloadFabricServer(params.Version, serverDir)
		default:
			emit("download", "active", "Downloading vanilla Java server jar", 42)
			err = DownloadJavaServer(params.Version, serverDir)
		}
	case models.EditionBedrock:
		emit("download", "active", "Downloading Bedrock dedicated server archive", 42)
		err = DownloadBedrockServer(params.Version, serverDir)
	}
	if err != nil {
		os.RemoveAll(serverDir)
		logger.Error(fmt.Sprintf("[CreateServer] Download failed id=%s: %s", id, err.Error()), nil)
		return nil, fmt.Errorf("download failed: %w", err)
	}
	logger.Info("[CreateServer] Download complete id="+id, nil)
	emit("download", "success", "Server artifact installed", 68)

	// Accept EULA for Java
	emit("configure", "active", "Writing server configuration", 78)
	if params.Edition == models.EditionJava {
		os.WriteFile(filepath.Join(serverDir, "eula.txt"), []byte("eula=true\n"), 0644)
	}

	// Write server.properties
	port := params.Port
	if port == 0 {
		if params.Edition == models.EditionJava {
			port = constants.DefaultJavaPort
		} else {
			port = constants.DefaultBedrockPort
		}
	}
	maxPlayers := params.MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = constants.DefaultMaxPlayers
	}
	memoryMB := params.MemoryMB
	if memoryMB == 0 {
		memoryMB = constants.DefaultMemoryMB
	}

	propPath := filepath.Join(serverDir, "server.properties")
	if params.Edition == models.EditionBedrock {
		// Bedrock: merge our settings into the extracted server.properties
		if existing, err := os.ReadFile(propPath); err == nil {
			merged := mergeProperties(string(existing), map[string]string{
				"server-port": fmt.Sprintf("%d", port),
				"max-players": fmt.Sprintf("%d", maxPlayers),
			})
			os.WriteFile(propPath, []byte(merged), 0644)
		} else {
			props := fmt.Sprintf("server-port=%d\nmax-players=%d\nmotd=MC Manage Server\n", port, maxPlayers)
			os.WriteFile(propPath, []byte(props), 0644)
		}
	} else {
		// Java: write a complete default server.properties
		props := fmt.Sprintf("server-port=%d\nmax-players=%d\nmotd=MC Manage Server\nonline-mode=true\ndifficulty=easy\ngamemode=survival\nview-distance=10\n", port, maxPlayers)
		os.WriteFile(propPath, []byte(props), 0644)
	}

	// Create plugins/addons dir
	emit("files", "active", "Preparing plugin directories", 88)
	subDir := "plugins"
	if params.Edition == models.EditionBedrock {
		subDir = "addons"
	}
	os.MkdirAll(filepath.Join(serverDir, subDir), os.ModePerm)

	// Create mods directory if using Forge or Fabric
	if serverType == "forge" || serverType == "fabric" {
		os.MkdirAll(filepath.Join(serverDir, "mods"), os.ModePerm)
	}

	config := &models.ServerConfig{
		ID:         id,
		Name:       params.Name,
		Edition:    params.Edition,
		ServerType: serverType,
		ModLoader:  serverType,
		Version:    params.Version,
		Port:       port,
		MaxPlayers: maxPlayers,
		MemoryMB:   memoryMB,
		Status:     models.StatusStopped,
		CreatedAt:  utils.NowISO(),
		ServerDir:  serverDir,
	}

	emit("register", "active", "Registering server", 94)
	state.AddServer(config)
	logger.Info("[CreateServer] Registered id="+config.ID+" name="+config.Name, nil)
	emit("complete", "success", "Server setup complete", 100, config.ID)
	return config, nil
}

func DeleteServer(state *AppState, id string) error {
	logger.Info("[DeleteServer] id="+id, nil)
	state.StopServer(id)

	srv, ok := state.RemoveServer(id)
	if ok {
		os.RemoveAll(srv.ServerDir)
		logger.Info("[DeleteServer] Removed dir id="+id, nil)
	}
	return nil
}

// ListPlugins returns the list of plugins/addons for a server
func ListPlugins(state *AppState, serverID string) ([]models.PluginInfo, error) {
	srv, ok := state.GetServer(serverID)
	if !ok {
		return nil, fmt.Errorf("server not found")
	}

	subDir := "plugins"
	if srv.Edition == models.EditionBedrock {
		subDir = "addons"
	}
	pluginsDir := filepath.Join(srv.ServerDir, subDir)

	var plugins []models.PluginInfo
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return plugins, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		plugins = append(plugins, models.PluginInfo{
			Name: entry.Name(),
			Size: info.Size(),
		})
	}
	return plugins, nil
}

// DeletePlugin removes a plugin file from the server
func DeletePlugin(state *AppState, serverID, name string) error {
	srv, ok := state.GetServer(serverID)
	if !ok {
		return fmt.Errorf("server not found")
	}

	// Sanitize filename
	safeName := filepath.Base(name)
	if safeName == "." || safeName == ".." {
		return fmt.Errorf("invalid filename")
	}

	subDir := "plugins"
	if srv.Edition == models.EditionBedrock {
		subDir = "addons"
	}
	pluginsDir := filepath.Join(srv.ServerDir, subDir)
	filePath := filepath.Join(pluginsDir, safeName)

	// Security: ensure path stays within plugins dir
	cleanPath := filepath.Clean(filePath)
	cleanDir := filepath.Clean(pluginsDir)
	if len(cleanPath) <= len(cleanDir) || cleanPath[:len(cleanDir)] != cleanDir {
		return fmt.Errorf("invalid filename")
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("plugin not found")
	}
	return nil
}

// DownloadJavaServer downloads a Java server jar for the given version
func DownloadJavaServer(version, dest string) error {
	logger.Info("[DownloadJavaServer] Fetching manifest for version="+version, nil)
	// Fetch version manifest
	resp, err := http.Get(constants.MojangVersionManifestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var manifest VersionManifest
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := jsonUnmarshal(body, &manifest); err != nil {
		return err
	}

	// Find version
	var versionURL string
	for _, v := range manifest.Versions {
		if v.ID == version {
			versionURL = v.URL
			break
		}
	}
	if versionURL == "" {
		logger.Warn("[DownloadJavaServer] Version not found: "+version, nil)
		return fmt.Errorf("version %s not found", version)
	}
	logger.Info("[DownloadJavaServer] Found URL for version="+version, nil)

	// Fetch version detail
	resp, err = http.Get(versionURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var detail VersionDetail
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := jsonUnmarshal(body, &detail); err != nil {
		return err
	}

	if detail.Downloads.Server == nil {
		logger.Warn("[DownloadJavaServer] No server download for version="+version, nil)
		return fmt.Errorf("no server download available")
	}
	logger.Info("[DownloadJavaServer] Downloading jar version="+version, nil)

	// Download server jar
	resp, err = http.Get(detail.Downloads.Server.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jarData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dest, "server.jar"), jarData, 0644)
}

// DownloadBedrockServer downloads and extracts a Bedrock server
func DownloadBedrockServer(version, dest string) error {
	osName := "linux"
	if runtime.GOOS == "windows" {
		osName = "win"
	}
	url := fmt.Sprintf("https://www.minecraft.net/bedrockdedicatedserver/bin-%s/bedrock-server-%s.zip", osName, version)
	logger.Info(fmt.Sprintf("[DownloadBedrockServer] Downloading version=%s os=%s", version, osName), nil)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	// Minecraft.net actively blocks Go's default User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/zip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return utils.ExtractZip(data, dest)
}

// DownloadPaperServer downloads a Paper server jar for the given version
func DownloadPaperServer(version, dest string) error {
	logger.Info("[DownloadPaperServer] Fetching builds for version="+version, nil)

	// Get latest build number
	buildsURL := fmt.Sprintf("https://api.papermc.io/v3/projects/paper/versions/%s/builds", version)
	resp, err := http.Get(buildsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Paper builds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Paper API returned status %d for version %s", resp.StatusCode, version)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var buildsResp struct {
		Builds []struct {
			Build     int `json:"build"`
			Downloads map[string]struct {
				Name string `json:"name"`
			} `json:"downloads"`
		} `json:"builds"`
	}
	if err := json.Unmarshal(body, &buildsResp); err != nil {
		return fmt.Errorf("failed to parse Paper builds: %w", err)
	}

	if len(buildsResp.Builds) == 0 {
		return fmt.Errorf("no Paper builds found for version %s", version)
	}

	latestBuild := buildsResp.Builds[len(buildsResp.Builds)-1]
	appDownload, ok := latestBuild.Downloads["application"]
	if !ok {
		return fmt.Errorf("no application download in Paper build")
	}

	jarURL := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%d/downloads/%s",
		version, latestBuild.Build, appDownload.Name)
	logger.Info(fmt.Sprintf("[DownloadPaperServer] Downloading build %d for version=%s", latestBuild.Build, version), nil)

	resp, err = http.Get(jarURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jarData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dest, "server.jar"), jarData, 0644)
}

// DownloadSpigotServer downloads a Spigot-compatible server jar for the given version.
// It tries the GetBukkit mirror first and falls back to Paper (which is Spigot-compatible
// and supports all Bukkit/Spigot plugins) on any error.
func DownloadSpigotServer(version, dest string) error {
	logger.Info("[DownloadSpigotServer] Downloading for version="+version, nil)

	// Try GetBukkit mirror
	url := fmt.Sprintf("https://cdn.getbukkit.org/spigot/spigot-%s.jar", version)
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		resp, doErr := http.DefaultClient.Do(req)
		if doErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if jarData, readErr := io.ReadAll(resp.Body); readErr == nil {
					// Verify it looks like a JAR (ZIP magic bytes PK)
					if len(jarData) >= 4 && jarData[0] == 0x50 && jarData[1] == 0x4B {
						return os.WriteFile(filepath.Join(dest, "server.jar"), jarData, 0644)
					}
					logger.Warn("[DownloadSpigotServer] GetBukkit file is not a valid JAR, falling back to Paper", nil)
				}
			} else {
				logger.Warn(fmt.Sprintf("[DownloadSpigotServer] GetBukkit returned HTTP %d, falling back to Paper", resp.StatusCode), nil)
			}
		} else {
			logger.Warn("[DownloadSpigotServer] GetBukkit unreachable: "+doErr.Error()+", falling back to Paper", nil)
		}
	}

	// Fallback: Paper is fully Spigot/Bukkit-compatible and supports all Spigot plugins
	logger.Info("[DownloadSpigotServer] Using Paper fallback for version="+version, nil)
	return DownloadPaperServer(version, dest)
}

// FetchJavaVersions fetches available Java edition versions from Mojang
func FetchJavaVersions() ([]models.VersionInfo, error) {
	resp, err := http.Get(constants.MojangVersionManifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest VersionManifest
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(body, &manifest); err != nil {
		return nil, err
	}

	var versions []models.VersionInfo
	count := 0
	for _, v := range manifest.Versions {
		if v.VersionType == "release" {
			versions = append(versions, models.VersionInfo{
				ID:          v.ID,
				VersionType: v.VersionType,
			})
			count++
			if count >= 30 {
				break
			}
		}
	}
	return versions, nil
}

// GetBedrockVersions returns the latest Bedrock version from the community API
func GetBedrockVersions() []models.VersionInfo {
	version, err := fetchLatestBedrockVersion()
	if err != nil {
		logger.Error("[GetBedrockVersions] Failed to fetch latest version: "+err.Error(), nil)
		// Fallback to empty or a default if necessary, but returning error info in models is better
		return []models.VersionInfo{}
	}

	return []models.VersionInfo{
		{
			ID:          version,
			VersionType: "release",
		},
	}
}

func fetchLatestBedrockVersion() (string, error) {
	resp, err := http.Get(constants.BedrockVersionAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiData BedrockVersionAPIResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &apiData); err != nil {
		return "", err
	}

	version := apiData.Linux.Stable
	if runtime.GOOS == "windows" {
		version = apiData.Windows.Stable
	}

	if version == "" {
		return "", fmt.Errorf("latest version not found in API response")
	}

	return version, nil
}

type BedrockVersionAPIResponse struct {
	Linux   BedrockOSVersion `json:"linux"`
	Windows BedrockOSVersion `json:"windows"`
}

type BedrockOSVersion struct {
	Stable string `json:"stable"`
}

// Internal types for Mojang API

type VersionManifest struct {
	Versions []ManifestVersion `json:"versions"`
}

type ManifestVersion struct {
	ID          string `json:"id"`
	VersionType string `json:"type"`
	URL         string `json:"url"`
}

type VersionDetail struct {
	Downloads Downloads `json:"downloads"`
}

type Downloads struct {
	Server *DownloadEntry `json:"server"`
}

type DownloadEntry struct {
	URL string `json:"url"`
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// DownloadForgeServer downloads the Forge installer and runs it to set up the server.
// Forge versions before 1.17 use a different setup, but for modern Forge (1.17+)
// we use the Forge installer to generate the server files.
func DownloadForgeServer(version, dest string) error {
	logger.Info("[DownloadForgeServer] Resolving Forge for version="+version, nil)

	// Step 1: Get the Forge version list from the Forge Maven metadata
	forgeVersion, err := resolveForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to resolve Forge version: %w", err)
	}

	logger.Info(fmt.Sprintf("[DownloadForgeServer] Resolved Forge version: %s", forgeVersion), nil)

	// Step 2: Download the Forge installer jar using the Maven format
	// Format: forge-{mc_version}-{forge_version}-installer.jar
	installerURL := fmt.Sprintf(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/forge-%s-%s-installer.jar",
		version, forgeVersion, version, forgeVersion,
	)
	logger.Info("[DownloadForgeServer] Downloading installer from: "+installerURL, nil)

	req, err := http.NewRequest("GET", installerURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download Forge installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Forge installer download failed: HTTP %d for version %s-%s", resp.StatusCode, version, forgeVersion)
	}

	installerPath := filepath.Join(dest, "forge-installer.jar")
	installerData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := os.WriteFile(installerPath, installerData, 0644); err != nil {
		return err
	}
	logger.Info("[DownloadForgeServer] Installer downloaded, running installServer", nil)

	// Step 3: Run the Forge installer with --installServer
	installCmd := exec.Command("java", "-jar", installerPath, "--installServer")
	installCmd.Dir = dest
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		logger.Warn("[DownloadForgeServer] Forge installer failed: "+err.Error(), nil)
		return fmt.Errorf("Forge installer failed: %w", err)
	}

	// Step 4: Remove the installer jar to clean up
	os.Remove(installerPath)

	// Step 5: Create mods directory
	os.MkdirAll(filepath.Join(dest, "mods"), os.ModePerm)

	logger.Info("[DownloadForgeServer] Forge server setup complete", nil)
	return nil
}

// resolveForgeVersion queries the Forge Maven metadata to find the latest Forge version
// for a given Minecraft version.
func resolveForgeVersion(mcVersion string) (string, error) {
	// The Forge maven metadata lists all available Forge versions for a Minecraft version
	mavenURL := fmt.Sprintf(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml",
	)
	resp, err := http.Get(mavenURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Forge maven metadata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse XML metadata to find versions matching mcVersion
	// The Forge version format is: mc_version-forge_version
	content := string(body)
	lines := strings.Split(content, "\n")
	var bestVersion string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<version>") && strings.HasSuffix(line, "</version>") {
			ver := strings.TrimPrefix(line, "<version>")
			ver = strings.TrimSuffix(ver, "</version>")
			// Check if this version is for our Minecraft version
			// Format: mc_version-forge_version or mc_version-forge_version-recommended
			if strings.HasPrefix(ver, mcVersion+"-") {
				// Prefer recommended builds
				if strings.Contains(ver, "recommended") {
					return ver, nil
				}
				// Keep the latest non-recommended as fallback
				parts := strings.Split(ver, "-")
				if len(parts) >= 2 {
					if bestVersion == "" || parts[1] > bestVersion {
						bestVersion = ver
					}
				}
			}
		}
	}

	if bestVersion == "" {
		// Fallback: try the latest version format (1.20.1-47.3.0 style)
		return mcVersion + "-latest", fmt.Errorf("no Forge version found for Minecraft %s; please check Forge availability", mcVersion)
	}

	return bestVersion, nil
}

// DownloadFabricServer downloads the Fabric server launcher.
// Fabric uses the Fabric Loader + Fabric API (optional) + vanilla server jar.
func DownloadFabricServer(version, dest string) error {
	logger.Info("[DownloadFabricServer] Resolving Fabric for version="+version, nil)

	// Step 1: Fetch the latest Fabric loader version for the Minecraft version
	loaderMetaURL := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s", version)
	resp, err := http.Get(loaderMetaURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Fabric loader versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fabric API returned HTTP %d for version %s", resp.StatusCode, version)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the JSON response to get the loader version
	var loaderData []struct {
		Loader struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
			Build   int    `json:"build"`
		} `json:"loader"`
	}
	if err := json.Unmarshal(body, &loaderData); err != nil {
		return fmt.Errorf("failed to parse Fabric loader data: %w", err)
	}

	if len(loaderData) == 0 || loaderData[0].Loader.Version == "" {
		return fmt.Errorf("no Fabric loader version found for Minecraft %s", version)
	}

	loaderVersion := loaderData[0].Loader.Version
	logger.Info(fmt.Sprintf("[DownloadFabricServer] Latest Fabric loader: %s", loaderVersion), nil)

	// Step 2: Get the Fabric installer version
	installerVerResp, err := http.Get("https://meta.fabricmc.net/v2/versions/installer")
	if err != nil {
		return fmt.Errorf("failed to fetch Fabric installer versions: %w", err)
	}
	defer installerVerResp.Body.Close()

	installerBody, err := io.ReadAll(installerVerResp.Body)
	if err != nil {
		return err
	}

	var installerVersions []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.Unmarshal(installerBody, &installerVersions); err != nil {
		return fmt.Errorf("failed to parse Fabric installer versions: %w", err)
	}

	installerVersion := "1.0.0"
	if len(installerVersions) > 0 {
		installerVersion = installerVersions[0].Version
	}

	// Step 3: Download the Fabric server launcher jar using the server launcher endpoint
	// The Fabric Meta API provides a direct server download URL
	serverURL := fmt.Sprintf(
		"https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
		version, loaderVersion, installerVersion,
	)
	logger.Info("[DownloadFabricServer] Downloading server launcher from: "+serverURL, nil)

	jarResp, err := http.Get(serverURL)
	if err != nil {
		return fmt.Errorf("failed to download Fabric server jar: %w", err)
	}
	defer jarResp.Body.Close()

	if jarResp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fabric server jar download failed: HTTP %d for version %s", jarResp.StatusCode, version)
	}

	jarData, err := io.ReadAll(jarResp.Body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dest, "fabric-server-launch.jar"), jarData, 0644); err != nil {
		return err
	}

	// Step 4: Create mods directory
	os.MkdirAll(filepath.Join(dest, "mods"), os.ModePerm)

	logger.Info("[DownloadFabricServer] Fabric server setup complete", nil)
	return nil
}

// ListMods returns the list of mod files for a Forge/Fabric server
func ListMods(state *AppState, serverID string) ([]models.ModInfo, error) {
	srv, ok := state.GetServer(serverID)
	if !ok {
		return nil, fmt.Errorf("server not found")
	}

	if srv.ModLoader != "forge" && srv.ModLoader != "fabric" {
		return nil, fmt.Errorf("mods are only available for Forge or Fabric servers")
	}

	modsDir := filepath.Join(srv.ServerDir, "mods")

	var mods []models.ModInfo
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return mods, nil // Return empty list if directory doesn't exist yet
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		mods = append(mods, models.ModInfo{
			Name: entry.Name(),
			Size: info.Size(),
		})
	}
	return mods, nil
}

// UploadMod saves an uploaded mod file into the server's mods directory
func UploadMod(state *AppState, serverID, filename string, reader io.Reader) error {
	srv, ok := state.GetServer(serverID)
	if !ok {
		return fmt.Errorf("server not found")
	}

	if srv.ModLoader != "forge" && srv.ModLoader != "fabric" {
		return fmt.Errorf("mods can only be uploaded to Forge or Fabric servers")
	}

	safeName := filepath.Base(filename)
	if safeName == "." || safeName == ".." {
		return fmt.Errorf("invalid filename")
	}

	modsDir := filepath.Join(srv.ServerDir, "mods")
	if err := os.MkdirAll(modsDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create mods directory: %w", err)
	}

	destPath := filepath.Join(modsDir, safeName)

	// Security: ensure path stays within mods dir
	cleanPath := filepath.Clean(destPath)
	cleanDir := filepath.Clean(modsDir)
	if len(cleanPath) <= len(cleanDir) || cleanPath[:len(cleanDir)] != cleanDir {
		return fmt.Errorf("invalid filename")
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// DeleteMod removes a mod file from the server's mods directory
func DeleteMod(state *AppState, serverID, name string) error {
	srv, ok := state.GetServer(serverID)
	if !ok {
		return fmt.Errorf("server not found")
	}

	if srv.ModLoader != "forge" && srv.ModLoader != "fabric" {
		return fmt.Errorf("mods can only be deleted from Forge or Fabric servers")
	}

	safeName := filepath.Base(name)
	if safeName == "." || safeName == ".." {
		return fmt.Errorf("invalid filename")
	}

	modsDir := filepath.Join(srv.ServerDir, "mods")
	filePath := filepath.Join(modsDir, safeName)

	// Security: ensure path stays within mods dir
	cleanPath := filepath.Clean(filePath)
	cleanDir := filepath.Clean(modsDir)
	if len(cleanPath) <= len(cleanDir) || cleanPath[:len(cleanDir)] != cleanDir {
		return fmt.Errorf("invalid filename")
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("mod not found")
	}
	return nil
}

// mergeProperties merges override values into an existing properties file content
func mergeProperties(content string, overrides map[string]string) string {
	lines := strings.Split(content, "\n")
	found := make(map[string]bool)
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}
		if idx := strings.IndexByte(trimmed, '='); idx >= 0 {
			key := strings.TrimSpace(trimmed[:idx])
			if val, ok := overrides[key]; ok {
				result = append(result, key+"="+val)
				found[key] = true
				continue
			}
		}
		result = append(result, line)
	}

	// Append any overrides not found in the original
	for key, val := range overrides {
		if !found[key] {
			result = append(result, key+"="+val)
		}
	}

	return strings.Join(result, "\n")
}
