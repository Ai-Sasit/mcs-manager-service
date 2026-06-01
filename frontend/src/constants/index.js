export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

export const WS_BASE_URL = API_BASE_URL.replace(/^http/, "ws").replace(
  /\/api\/v1$/,
  "",
);

export const API_TIMEOUT = 240000;

export const ENDPOINTS = {
  SERVERS: "/servers",
  JAVA_VERSIONS: "/versions/java",
  BEDROCK_VERSIONS: "/versions/bedrock",
};

export const DEFAULT_JAVA_PORT = 25565;
export const DEFAULT_BEDROCK_PORT = 19132;
export const DEFAULT_MAX_PLAYERS = 20;
export const DEFAULT_MEMORY_MB = 1024;