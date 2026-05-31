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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

type CreateServerParams struct {
	Name       string
	Edition    models.ServerEdition
	ServerType string
	Version    string
	Port       uint16
	MaxPlayers uint32
	MemoryMB   uint32
}

func CreateServer(state *AppState, params CreateServerParams) (*models.ServerConfig, error) {
	id := uuid.New().String()
	serverDir := filepath.Join(state.DataDir, "servers", id)
	logger.Info(fmt.Sprintf("[CreateServer] id=%s name=%s edition=%s version=%s", id, params.Name, params.Edition, params.Version), nil)

	if err := os.MkdirAll(serverDir, os.ModePerm); err != nil {
		logger.Error("[CreateServer] Failed to create dir: "+err.Error(), nil)
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	// Download server files
	var err error
	serverType := params.ServerType
	if serverType == "" {
		serverType = "vanilla"
	}
	logger.Info(fmt.Sprintf("[CreateServer] Downloading %s server %s type=%s", params.Edition, params.Version, serverType), nil)
	switch params.Edition {
	case models.EditionJava:
		switch serverType {
		case "paper":
			err = DownloadPaperServer(params.Version, serverDir)
		case "spigot":
			err = DownloadSpigotServer(params.Version, serverDir)
		default:
			err = DownloadJavaServer(params.Version, serverDir)
		}
	case models.EditionBedrock:
		err = DownloadBedrockServer(params.Version, serverDir)
	}
	if err != nil {
		os.RemoveAll(serverDir)
		logger.Error(fmt.Sprintf("[CreateServer] Download failed id=%s: %s", id, err.Error()), nil)
		return nil, fmt.Errorf("download failed: %w", err)
	}
	logger.Info("[CreateServer] Download complete id="+id, nil)

	// Accept EULA for Java
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
	subDir := "plugins"
	if params.Edition == models.EditionBedrock {
		subDir = "addons"
	}
	os.MkdirAll(filepath.Join(serverDir, subDir), os.ModePerm)

	config := &models.ServerConfig{
		ID:         id,
		Name:       params.Name,
		Edition:    params.Edition,
		ServerType: serverType,
		Version:    params.Version,
		Port:       port,
		MaxPlayers: maxPlayers,
		MemoryMB:   memoryMB,
		Status:     models.StatusStopped,
		CreatedAt:  utils.NowISO(),
		ServerDir:  serverDir,
	}

	state.AddServer(config)
	logger.Info("[CreateServer] Registered id="+config.ID+" name="+config.Name, nil)
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
	buildsURL := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds", version)
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
